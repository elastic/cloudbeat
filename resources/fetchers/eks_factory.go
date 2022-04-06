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
	awsConfigProvider := aws.ConfigProvider{}
	awsConfig := awsConfigProvider.GetConfig()
	eks := aws.NewEksProvider(awsConfig.Config)

	manager.Factories.ListFetcherFactory(EKSType, &EKSFactory{
		eksProvider: eks,
	})
}

type EKSFactory struct {
	eksProvider aws.EksClusterDescriber
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
	fe := &EKSFetcher{
		cfg:         cfg,
		eksProvider: f.eksProvider,
	}

	return fe, nil
}
