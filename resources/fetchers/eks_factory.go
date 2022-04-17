package fetchers

import (
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/providers/awslib"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	EKSType = "aws-eks"
)

func init() {
	manager.Factories.ListFetcherFactory(EKSType, &EKSFactory{
		extraElements: getEksExtraElements,
	})
}

type EKSFactory struct {
	extraElements func() (eksExtraElements, error)
}

type eksExtraElements struct {
	eksProvider awslib.EksClusterDescriber
}

func (f *EKSFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	logp.L().Info("EKS factory has started")
	cfg := EKSFetcherConfig{}
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

func getEksExtraElements() (eksExtraElements, error) {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig, err := awsConfigProvider.GetConfig()
	if err != nil {
		return eksExtraElements{}, err
	}

	eks := awslib.NewEksProvider(awsConfig.Config)

	return eksExtraElements{eksProvider: eks}, nil
}

func (f *EKSFactory) CreateFrom(cfg EKSFetcherConfig, elements eksExtraElements) (fetching.Fetcher, error) {
	fe := &EKSFetcher{
		cfg:         cfg,
		eksProvider: elements.eksProvider,
	}

	return fe, nil
}
