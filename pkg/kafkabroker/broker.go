package kafkabroker

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

type KafkaStruct struct {
	AdminClient *kafka.AdminClient
	L           *sync.RWMutex
}

var once sync.Once
var adminKafka KafkaStruct

// TODO: Broker connection close logic
//defer func() {
//	if err := kafkabroker.Close(); err != nil {
//		panic(err)
//	}
//}()
func Initialize() {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "192.168.130.7",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		logrus.Error(err)
		logrus.Debug("Fail to load-balancer kafkabroker connection")
	} else {
		adminKafka.AdminClient = admin
	}
}

func GetInstance() *KafkaStruct {
	once.Do(func() {
		Initialize()
	})
	return &adminKafka
}

func (k *KafkaStruct) GetAllTopics() ([]string) {
	topics := []string{}
	getTopics, err := k.AdminClient.GetMetadata(nil, true, 100)
	if err != nil {
		logrus.Error(err)
		logrus.Debug("Fail to get all topics list")
		return nil
	} else {
		for topic, _ := range getTopics.Topics {
			topics = append(topics, topic)
		}
		return topics
	}
}

func (k *KafkaStruct) DeleteTopics(topic string) error {
	topics := []string{topic}
	_, err := k.AdminClient.DeleteTopics(context.Background(), topics)
	if err != nil {
		logrus.Error(err)
		logrus.Debug("Fail to delete topic list from broker")
		return err
	}
	return nil
}
