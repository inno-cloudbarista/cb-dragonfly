package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
)

type ClientOptions struct {
	Endpoints string
}

type Config struct {
	ClientOptions ClientOptions
}

type Storage struct {
	Config Config
	Client client.Client
	L      *sync.RWMutex
}

var once sync.Once
var storage Storage

func Initialize() error {
	cfg := client.Config{
		Endpoints: []string{
			config.GetInstance().GetETCDConfig().EndpointUrl,
		},
	}
	if client, err := client.New(cfg); err != nil {
		logrus.Error(err)
		return err
	} else {
		storage.Client = client
	}
	storage.L = &sync.RWMutex{}
	return nil
}

func GetInstance() *Storage {
	once.Do(func() {
		Initialize()
	})
	return &storage
}

func (s *Storage) Init() error {
	return nil
}

func (s *Storage) WriteMetric(key string, metric interface{}, useTTL bool) error {

	kapi := client.NewKeysAPI(s.Client)
	var metricVal string

	_, ok := metric.(map[string]interface{})
	if ok {

		s.L.Lock()
		bytes, err := json.Marshal(metric)
		s.L.Unlock()

		if err != nil {
			logrus.Error("Failed to marshaling realtime monitoring data to JSON: ", err)
			return err
		}

		metricVal = fmt.Sprintf("%s", bytes)
	} else {
		metricVal = metric.(string)
	}

	var opts client.SetOptions
	if useTTL {
		opts = client.SetOptions{TTL: time.Duration(config.GetInstance().Monitoring.AgentInterval+1) * time.Second}
	} else {
		opts = client.SetOptions{TTL: -1}

	}

	s.L.RLock()
	_, err := kapi.Set(context.Background(), key, fmt.Sprintf("%s", metricVal), &opts)
	s.L.RUnlock()
	if err != nil {
		logrus.Error("Failed to write realtime monitoring data to ETCD : ", err)
		return err
	}
	return nil
}

//func (s *Storage) ReadMetric(key string) (map[string]interface{}, error) {
func (s *Storage) ReadMetric(key string) (*client.Node, error) {

	kapi := client.NewKeysAPI(s.Client)
	// 실시간 모니터링 데이터 조회
	resp, err := kapi.Get(context.Background(), key, nil)
	if err != nil {
		logrus.Error("Failed to read realtime monitoring data to ETCD : ", err)
		return nil, err
	}

	if resp == nil {
		//	s.L.RUnlock()
		return nil, nil
	}

	return resp.Node, nil
}

func (s *Storage) DeleteMetric(key string) error {
	kapi := client.NewKeysAPI(s.Client)

	// 실시간 모니터링 데이터 삭제
	opts := client.DeleteOptions{Recursive: true}
	_, err := kapi.Delete(context.Background(), key, &opts)
	if err != nil {
		logrus.Error("Failed to delete realtime monitoring data to ETCD : ", err)
		return err
	}

	return nil
}

func (s *Storage) WatchETCDSet(key string, ch *chan string) {
	kapi := client.NewKeysAPI(s.Client)
	watcher := kapi.Watcher(key, &client.WatcherOptions{Recursive: true})

	for true {
		resp, err := watcher.Next(context.TODO())
		if err != nil {
			if _, ok := err.(*client.ClusterError); ok {
				continue
			}
			logrus.Error(err)
		}
		switch resp.Action {
		case "set":
			topics := resp.Node.Value
			*ch <- topics
		}
	}
}


func (s *Storage) WatchETCDExpire(key string, ch *chan string) {
	kapi := client.NewKeysAPI(s.Client)
	watcher := kapi.Watcher(key, &client.WatcherOptions{Recursive: true})

	for true {
		resp, err := watcher.Next(context.TODO())
		if err != nil {
			if _, ok := err.(*client.ClusterError); ok {
				continue
			}
			logrus.Error(err)
		}
		switch resp.Action {
		case "expire":
			topic := resp.PrevNode.Value
			*ch <- topic
		}
	}
}
