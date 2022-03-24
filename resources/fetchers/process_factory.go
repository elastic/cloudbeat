package fetchers

import (
	"encoding/gob"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
	"os"
)

const (
	ProcessType = "process"
)

type ProcessFactory struct {
}

func init() {
	manager.Factories.ListFetcherFactory(ProcessType, &ProcessFactory{})
	gob.Register(ProcessResource{})
	gob.Register(map[string]interface{}{})
}

func (f *ProcessFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := ProcessFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	cfg.Fs = os.DirFS(cfg.Directory)
	return f.CreateFrom(cfg)
}

func (f *ProcessFactory) CreateFrom(cfg ProcessFetcherConfig) (fetching.Fetcher, error) {
	fe := &ProcessesFetcher{
		cfg: cfg,
	}

	return fe, nil
}
