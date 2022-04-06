package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"log"
)

type ConfigGetter interface {
	GetConfig() Config
}

type ConfigProvider struct {
}

func (p ConfigProvider) GetConfig() Config {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		Config: cfg,
	}
}
