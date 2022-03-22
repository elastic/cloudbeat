package fetchers

import (
	"encoding/gob"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	EKSType = "aws-eks"
)

func init() {
	manager.Factories.ListFetcherFactory(EKSType, &EKSFactory{})
	gob.Register(EKSResource{})
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
	fe := &EKSFetcher{
		cfg: cfg,
	}

	return fe, nil
}

func NewEKSFetcher(awsCfg AwsFetcherConfig, cfg EKSFetcherConfig) (fetching.Fetcher, error) {
	eks := NewEksProvider(awsCfg.Config)

	return &EKSFetcher{
		cfg:         cfg,
		eksProvider: eks,
	}, nil
}
