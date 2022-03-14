package fetchers

import (
	"github.com/elastic/beats/v7/libbeat/common"
)

type ProcessFactory struct {
}

func (f *ProcessFactory) Create(c common.Config) (Fetcher, error) {
	cfg := ProcessFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *ProcessFactory) CreateFrom(cfg ProcessFetcherConfig) (Fetcher, error) {
	fe := &ProcessesFetcher{
		cfg: cfg,
	}

	return fe, nil
}
