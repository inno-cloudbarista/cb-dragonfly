package alert

import (
	"time"

	kclient "github.com/shaodan/kapacitor-client"

	"github.com/cloud-barista/cb-dragonfly/pkg/config"
)

const (
	kapacitorTimeout = 300
)

//var once sync.Once
//var client *kclient.Client

func newClient() (*kclient.Client, error) {
	kapacitorConfig := kclient.Config{
		URL:                config.GetDefaultConfig().GetKapacitorConfig().GetEndpointUrl(),
		Timeout:            time.Duration(kapacitorTimeout),
		InsecureSkipVerify: true,
	}
	client, err := kclient.New(kapacitorConfig)
	return client, err
}

func GetClient() *kclient.Client {
	c, _ := newClient()
	return c
}