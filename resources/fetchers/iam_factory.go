package fetchers

import (
	"github.com/elastic/cloudbeat/resources/providers/awslib"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	IAMType = "aws-iam"
)

func init() {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig := awsConfigProvider.GetConfig()
	provider := awslib.NewIAMProvider(awsConfig.Config)

	manager.Factories.ListFetcherFactory(IAMType, &IAMFactory{
		iamProvider: provider,
	})
}

type IAMFactory struct {
	iamProvider awslib.IAMRolePermissionGetter
}

func (f *IAMFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := IAMFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *IAMFactory) CreateFrom(cfg IAMFetcherConfig) (fetching.Fetcher, error) {
	return &IAMFetcher{
		cfg:         cfg,
		iamProvider: f.iamProvider,
	}, nil

}
