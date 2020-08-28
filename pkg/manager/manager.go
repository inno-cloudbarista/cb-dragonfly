package manager

import (
	"errors"
	"fmt"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"

	"github.com/cloud-barista/cb-dragonfly/pkg/collector"
	"github.com/cloud-barista/cb-dragonfly/pkg/config"
	"github.com/cloud-barista/cb-dragonfly/pkg/metricstore/influxdbv1"
	"github.com/cloud-barista/cb-dragonfly/pkg/realtimestore/etcd"
)

// TODO: implements
// TODO: 1. API Server
// TODO: 2. Scheduling Collector...
// TODO: 3. Configuring Policy...

type CollectManager struct {
	CollectorManageSlice [] *collector.MetricCollector
	WaitGroup *sync.WaitGroup
	metricL   *sync.RWMutex
}

func NewCollectorManager() (*CollectManager, error) {
	manager := CollectManager{}

	influxConfig := influxdbv1.Config{
		ClientOptions: []influxdbv1.ClientOptions{
			{
				URL:      config.GetInstance().GetInfluxDBConfig().EndpointUrl,
				Username: config.GetInstance().GetInfluxDBConfig().UserName,
				Password: config.GetInstance().GetInfluxDBConfig().Password,
			},
		},
		Database: config.GetInstance().GetInfluxDBConfig().Database,
	}

	err := influxdbv1.Initialize(influxConfig)
	if err != nil {
		logrus.Error("Failed to initialize influxDB")
		return nil, err
	}

	manager.metricL = &sync.RWMutex{}

	return &manager, nil
}

func (manager *CollectManager) FlushMonitoringData() {
	etcd.GetInstance().DeleteMetric("/collector")
	etcd.GetInstance().DeleteMetric("/host")
	etcd.GetInstance().DeleteMetric("/mon")
	manager.SetConfigurationToETCD()
}

func (manager *CollectManager) SetConfigurationToETCD() {
	monConfigMap := map[string]interface{}{}
	mapstructure.Decode(config.GetInstance().Monitoring, &monConfigMap)
	etcd.GetInstance().WriteMetric("/mon/config", monConfigMap, false)
}

func (manager *CollectManager) StartScheduler(wg *sync.WaitGroup) error {
	defer wg.Done()
	scheduler, err := NewCollectorScheduler(manager)
	if err != nil {
		logrus.Error("Failed to initialize influxDB")
		return err
	}
	go scheduler.Scheduler()
	return nil
}

func (manager *CollectManager) StartCollector(wg *sync.WaitGroup) error {
	manager.WaitGroup = wg
	test := config.GetInstance()
	fmt.Println(test)
	for i := 0; i < config.GetInstance().CollectManager.CollectorCnt; i++ {
		err := manager.CreateCollector()
		if err != nil {
			logrus.Error("failed to create collector", err)
			continue
		}
	}

	return nil
}

func (manager *CollectManager) CreateCollector() error {

	fmt.Println("Create Collector")
	collectorIdx := len(manager.CollectorManageSlice)

	mc := collector.NewMetricCollector(
		manager.metricL,
		collector.AVG,
		collectorIdx,
	)
	manager.CollectorManageSlice = append(manager.CollectorManageSlice, &mc)

	manager.WaitGroup.Add(1)
	go func() {
		collectorUUID := fmt.Sprintf("/collector/%d/uuid/%s", mc.CreateOrder, mc.UUID)
		err := etcd.GetInstance().WriteMetric(collectorUUID,  mc.UUID, true)
		if err != nil {
			logrus.Error("fail to write collectorUUID to ETCD", err)
		}
		err = mc.Collector(manager.WaitGroup)
		if err != nil {
			logrus.Debug("Fail to create Collector")
		}
	}()

	return nil
}

func (manager *CollectManager) StopCollector() error {
	if len(manager.CollectorManageSlice) == 1 {
		return nil
	}
	collectorIdx := len(manager.CollectorManageSlice)-1
	manager.CollectorManageSlice[collectorIdx].Active = false
	manager.CollectorManageSlice = manager.CollectorManageSlice[:collectorIdx]
	err := etcd.GetInstance().DeleteMetric(fmt.Sprintf("/collector/%d", collectorIdx))
	if err != nil {
		return errors.New(fmt.Sprintf("failed to get collector by collectorIdx: %d", collectorIdx))
	}
	return nil
}
