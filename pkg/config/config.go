package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	InfluxDB       InfluxDB
	CollectManager CollectManager
	APIServer      APIServer
	Monitoring     Monitoring
	Kapacitor      Kapacitor
}

type InfluxDB struct {
	EndpointUrl  string `json:"endpoint_url" mapstructure:"endpoint_url"`
	InternalPort int    `json:"internal_port" mapstructure:"internal_port"`
	ExternalPort int    `json:"external_port" mapstructure:"external_port"`
	Database     string
	UserName     string `json:"user_name" mapstructure:"user_name"`
	Password     string
}

type CollectManager struct {
	CollectorIP             string `json:"collector_ip" mapstructure:"collector_ip"`
	CollectorPort int    `json:"collector_port" mapstructure:"collector_port"`
	CollectorGroupCnt       int    `json:"collectorGroup_count" mapstructure:"collectorGroup_count"`
	//GroupPerCollectCnt int    `json:"group_per_collect_count" mapstructure:"group_per_collect_count"`
}

type APIServer struct {
	Port int
}

type Monitoring struct {
	AgentInterval      int `json:"agent_interval" mapstructure:"agent_interval"`         // 모니터링 에이전트 수집주기
	CollectorInterval  int `json:"collector_interval" mapstructure:"collector_interval"` // 모니터링 콜렉터 Aggregate 주기
	MaxHostCount       int `json:"max_host_count" mapstructure:"max_host_count"`         // 모니터링 콜렉터 수
	MonitoringPolicy       int `json:"monitoring_policy" mapstructure:"monitoring_policy"`         // 모니터링 콜렉터 수
}

type Kapacitor struct {
	EndpointUrl string `json:"endpoint_url" mapstructure:"endpoint_url"`
}

func (kapacitor Kapacitor) GetEndpointUrl() string {
	return kapacitor.EndpointUrl
}

var once sync.Once
var config Config

func GetInstance() *Config {
	once.Do(func() {
		loadConfigFromYAML(&config)
	})
	return &config
}

func GetDefaultConfig() *Config {
	var defaultMonConfig Config
	loadConfigFromYAML(&defaultMonConfig)
	return &defaultMonConfig
}

func (config *Config) SetMonConfig(newMonConfig Monitoring) {
	config.Monitoring = newMonConfig
}

func (config *Config) GetInfluxDBConfig() InfluxDB {
	return config.InfluxDB
}

func (config *Config) GetKapacitorConfig() Kapacitor {
	return config.Kapacitor
}

func loadConfigFromYAML(config *Config) {
	configPath := os.Getenv("CBMON_ROOT") + "/conf"

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}
