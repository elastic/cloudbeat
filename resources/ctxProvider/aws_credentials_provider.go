package ctxProvider

import (
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"log"
)

type AWSCredProvider struct {
}

func (p AWSCredProvider) GetAwsCredentials() AwsFetcherConfig {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		log.Fatal(err)
	}

	return AwsFetcherConfig{
		Config: cfg,
	}
}
