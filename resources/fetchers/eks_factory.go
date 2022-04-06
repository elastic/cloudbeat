package fetchers

import (
	"github.com/elastic/cloudbeat/resources/providers/aws"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	EKSType = "aws-eks"
)

func init() {
	manager.Factories.ListFetcherFactory(EKSType, &EKSFactory{})
}

type EKSFactory struct {
}

func (f *EKSFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := EKSFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *EKSFactory) CreateFrom(cfg EKSFetcherConfig) (fetching.Fetcher, error) {
	awsCredProvider := aws.AWSCredProvider{}
	awsCfg := awsCredProvider.GetAwsCredentials()
	eks := aws.NewEksProvider(awsCfg.Config)

	fe := &EKSFetcher{
		cfg:         cfg,
		eksProvider: eks,
	}

	return fe, nil
}
