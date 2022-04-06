package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"log"
)

type AwsCredentialsGetter interface {
	GetAwsCredentials() Config
}

type AWSCredProvider struct {
}

func (p AWSCredProvider) GetAwsCredentials() Config {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		Config: cfg,
	}
}
