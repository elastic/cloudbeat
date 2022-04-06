package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"log"
)

type AwsCredentialsGetter interface {
	GetAwsCredentials() FetcherConfig
}

type AWSCredProvider struct {
}

func (p AWSCredProvider) GetAwsCredentials() FetcherConfig {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Fatal(err)
	}

	return FetcherConfig{
		Config: cfg,
	}
}
