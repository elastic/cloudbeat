package fetchers

import (
	"encoding/gob"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	ELBType = "aws-elb"
)

func init() {
	manager.Factories.ListFetcherFactory(ELBType, &ELBFactory{})
	gob.Register(ELBResource{})
}

type ELBFactory struct {
}

func (f *ELBFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := ELBFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *ELBFactory) CreateFrom(cfg ELBFetcherConfig) (fetching.Fetcher, error) {
	fe := &ELBFetcher{
		cfg: cfg,
	}

	return fe, nil
}
