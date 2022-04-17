package fetchers

import (
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/providers/awslib"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	IAMType = "aws-iam"
)

func init() {

	manager.Factories.ListFetcherFactory(IAMType, &IAMFactory{
		extraElements: getIamExtraElements,
	})
}

type IAMFactory struct {
	extraElements func() (IAMExtraElements, error)
}

type IAMExtraElements struct {
	iamProvider awslib.IAMRolePermissionGetter
}

func (f *IAMFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	logp.L().Info("IAM factory has started")
	cfg := IAMFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}
	elements, err := f.extraElements()
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg, elements)
}

func getIamExtraElements() (IAMExtraElements, error) {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig, err := awsConfigProvider.GetConfig()
	if err != nil {
		return IAMExtraElements{}, err
	}
	provider := awslib.NewIAMProvider(awsConfig.Config)

	return IAMExtraElements{
		iamProvider: provider,
	}, nil
}

func (f *IAMFactory) CreateFrom(cfg IAMFetcherConfig, elements IAMExtraElements) (fetching.Fetcher, error) {
	return &IAMFetcher{
		cfg:         cfg,
		iamProvider: elements.iamProvider,
	}, nil

}
