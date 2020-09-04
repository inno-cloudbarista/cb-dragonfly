package collector

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/kafkabroker"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore/etcd"
)

type MetricCollector struct {
	UUID              string
	CreateOrder int
	KafkaConn		  *kafka.Consumer
	metricL           *sync.RWMutex
	Aggregator        Aggregator
	Active           bool
}

type TelegrafMetric struct {
	Name      string                 `json:"name"`
	Tags      map[string]interface{} `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
	Timestamp int64                  `json:"timestamp"`
	TagInfo   map[string]interface{} `json:"tagInfo"`
}

func NewMetricCollector(mutexLock *sync.RWMutex, aggregateType AggregateType, createOrder int) MetricCollector {

	uuid := uuid.New().String()
	kafkaConn, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "192.168.130.7",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		logrus.Error("Fail to create collector kafkabroker consumer", err)
		logrus.Debug(err)
	}

	mc := MetricCollector{
		UUID:              uuid,
		KafkaConn:		   kafkaConn,
		CreateOrder: createOrder,
		metricL:           mutexLock,
		Aggregator: Aggregator{
			AggregateType: aggregateType,
		},
		Active:           true,
	}

	return mc
}

func (mc *MetricCollector) Collector(wg *sync.WaitGroup) error {

	defer wg.Done()
	var initialSubscribe sync.Once

	ch := make(chan string)
	go func() {
		processDecision := "aggregateMetric"
		for {
			collectorInterval := config.GetInstance().Monitoring.CollectorInterval
			time.Sleep(time.Duration(collectorInterval) * time.Second)
			ch <- processDecision
		}
	}()

	for {
		select {
		case processDecision := <-ch:
			switch {
			case processDecision == "aggregateMetric":
				logrus.Debug("======================================================================")
				logrus.Debug("[" + mc.UUID + "]Start Aggregate!!")
				// TODO: Consider goroutine based aggregating logic
				// Current design pattern => During metric aggregating, the collector doesnt collect metric.
				// if we use above the goroutine based aggregating logic, than we should consider that etcd con-currency issue occur
				err := mc.Aggregator.AggregateMetric(mc.UUID)
				if err != nil {
					logrus.Error("["+mc.UUID+"]Failed to aggregate meric", err)
				}
				logrus.Debug("======================================================================")
			default:
				fmt.Println("===processDecision ===")
				fmt.Println(processDecision)
				if string(processDecision[0]) == "&" {
					topics := strings.Split(processDecision, "&")[1:]
					fmt.Println(topics)
					mc.KafkaConn.SubscribeTopics(topics, nil)
					for _, topic := range topics {
						topicInfo := fmt.Sprintf("/collector/%d/topics/%s", mc.CreateOrder,  topic)
						err := etcd.GetInstance().WriteMetric(topicInfo, topic, true)
						if err != nil {
							logrus.Error("fail to write topicInfo to ETCD", err)
							return nil
						}
					}
				} else {
					topic := processDecision
					fmt.Println("=== im deleted Topic ===")
					fmt.Println(topic)
					_ = kafkabroker.GetInstance().DeleteTopics(topic)
				}
			}
		default:
			initialSubscribe.Do(func() {
				go etcd.GetInstance().WatchETCDSet(fmt.Sprintf("topicschedule/%d/topics",mc.CreateOrder), &ch)
				go etcd.GetInstance().WatchETCDExpire(fmt.Sprintf("collector/%d/topics", mc.CreateOrder), &ch)
			})

			if !mc.Active {
				fmt.Println("Deleting Collector!")
				close(ch)
				err := mc.KafkaConn.Close()
				if err != nil {
					logrus.Debug("Fail to collector kafkabroker connection close")
				}
				return nil
			} else {
				agentInterval := config.GetInstance().Monitoring.AgentInterval
				msg, err := mc.KafkaConn.ReadMessage((time.Duration(agentInterval)+1)* time.Second)
				if err != nil {
					logrus.Error("fail to read collector kafkabroker msg", err)
				}

				if msg != nil {
					_ = mc.cashingMetricToETCD(msg)
				}
			}
		}
	}
}

func (mc *MetricCollector) cashingMetricToETCD(msg *kafka.Message) error {
	metric := TelegrafMetric{}
	err := json.Unmarshal(msg.Value, &metric)
	var hostId string
	if hostIdVal, ok := metric.Tags["hostID"].(string); !ok {
		return nil
	} else {
		hostId = hostIdVal
	}

	topicInfo := fmt.Sprintf("/collector/%d/topics/%s", mc.CreateOrder,  *msg.TopicPartition.Topic)
	err = etcd.GetInstance().WriteMetric(topicInfo, *msg.TopicPartition.Topic, true)
	if err != nil {
		logrus.Error("fail to write topicInfo to ETCD", err)
		return nil
	}

	hostInfo := fmt.Sprintf("/collector/%d/host/%s", mc.CreateOrder, hostId)
	err = etcd.GetInstance().WriteMetric(hostInfo, hostId, true)
	if err != nil {
		logrus.Error("fail to write hostInfo to ETCD", err)
		return nil
	}

	curTimestamp := time.Now().Unix()
	var diskName string
	var metricKey string
	var osTypeKey string

	mc.metricL.RLock()
	switch strings.ToLower(metric.Name) {
	case "disk":
		diskName = metric.Tags["device"].(string)
		diskName = strings.ReplaceAll(diskName, "/", "%")
		metricKey = fmt.Sprintf("/host/%s/metric/%s/%s/%d", hostId, metric.Name, diskName, curTimestamp)
	case "diskio":
		diskName := metric.Tags["name"].(string)
		diskName = strings.ReplaceAll(diskName, "/", "%")
		metricKey = fmt.Sprintf("/host/%s/metric/%s/%s/%d", hostId, metric.Name, diskName, curTimestamp)
	default:
		metricKey = fmt.Sprintf("/host/%s/metric/%s/%d", hostId, metric.Name, curTimestamp)
	}
	mc.metricL.RUnlock()

	if err := etcd.GetInstance().WriteMetric(metricKey, metric.Fields, true); err != nil {
		logrus.Error(err)
	}

	metric.TagInfo = map[string]interface{}{}
	metric.TagInfo["mcisId"] = hostId
	metric.TagInfo["hostId"] = hostId
	metric.TagInfo["osType"] = metric.Tags["osType"].(string)

	osTypeKey = fmt.Sprintf("/host/%s/tag", hostId)

	if err := etcd.GetInstance().WriteMetric(osTypeKey, metric.TagInfo, true); err != nil {
		logrus.Error(err)
	}
	return nil
}
