package collector

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type MetricCollector struct {
	CreateOrder        int
	ConsumerKafkaConn  *kafka.Consumer
	AdminKafkaConn     *kafka.AdminClient
	Aggregator         Aggregator
	Active             bool
	Ch                 chan string
}

type TelegrafMetric struct {
	Name      string                 `json:"name"`
	Tags      map[string]interface{} `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
	Timestamp int64                  `json:"timestamp"`
	TagInfo   map[string]interface{} `json:"tagInfo"`
}

var KafkaConfig *kafka.ConfigMap

func NewMetricCollector(aggregateType AggregateType, createOrder int) MetricCollector {

	KafkaConfig = &kafka.ConfigMap{
		"bootstrap.servers":  "192.168.130.7",
		"group.id":           fmt.Sprintf("%d", createOrder),
		"enable.auto.commit": true,
		"auto.offset.reset":  "earliest",
	}

	consumerKafkaConn, err := kafka.NewConsumer(KafkaConfig)
	if err != nil {
		logrus.Error("Fail to create collector kafka consumer", err)
		logrus.Debug(err)
	}
	adminKafkaConn, err := kafka.NewAdminClient(KafkaConfig)
	if err != nil {
		logrus.Error("Fail to create collector kafka consumer", err)
		logrus.Debug(err)
	}
	ch := make(chan string)
	mc := MetricCollector{
		ConsumerKafkaConn: consumerKafkaConn,
		AdminKafkaConn:    adminKafkaConn,
		CreateOrder:       createOrder,
		Aggregator: Aggregator{
			AggregateType: aggregateType,
		},
		Active:             true,
		Ch:                 ch,
	}
	return mc
}

func (mc *MetricCollector) Collector(wg *sync.WaitGroup) error {

	defer wg.Done()
	DeliveredTopicList := []string {}
	currentSubscribeTopicList := []string{}
	aliveTopics := []string{}
	getTopicsAllow := true
	for {
		select {
		case processDecision := <-mc.Ch:
			if len(processDecision) != 0 {
				currentSubscribeTopicList, _ = mc.ConsumerKafkaConn.Subscription()
				sort.Strings(currentSubscribeTopicList)
				DeliveredTopicList = unique(strings.Split(processDecision, "&")[1:])
				fmt.Println(fmt.Sprintf("Group_%d collector Delivered : %s", mc.CreateOrder, DeliveredTopicList))
				if !cmp.Equal(DeliveredTopicList, currentSubscribeTopicList) && getTopicsAllow {
					_ = mc.ConsumerKafkaConn.SubscribeTopics(DeliveredTopicList, nil)
				}
				if !getTopicsAllow {
					DeliveredTopicList = aliveTopics
					getTopicsAllow = true
				}
				aliveTopics, _ = mc.Aggregator.AggregateMetric(mc.ConsumerKafkaConn, DeliveredTopicList)
			}
			break
		}
		if !cmp.Equal(aliveTopics, DeliveredTopicList) {
			if len(aliveTopics) == 0 {
				_ = mc.ConsumerKafkaConn.Unsubscribe()
			} else {
				_ = mc.ConsumerKafkaConn.SubscribeTopics(aliveTopics, nil)
			}
			_, _ = mc.AdminKafkaConn.DeleteTopics(context.Background(), ReturnDiffTopicList(DeliveredTopicList, aliveTopics))
			getTopicsAllow = false
		}
		if !mc.Active {
			close(mc.Ch)
			err := mc.ConsumerKafkaConn.Close()
			if err != nil {
				logrus.Debug("Fail to collector kafka connection close")
			}
			mc.AdminKafkaConn.Close()
			return nil
		}
	}
}
