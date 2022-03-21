package fetchers

import (
	"encoding/gob"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
	"os"
	"strings"
)

const (
	ProcessType = "process"
)

type ProcessFactory struct {
}

func init() {
	manager.Factories.ListFetcherFactory(ProcessType, &ProcessFactory{})
	gob.Register(ProcessResource{})
}

func (f *ProcessFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := ProcessFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(cfg.Directory, "/hostfs") {
		return nil, fmt.Errorf("process fetcher could not start - the directory path should start with `/hostfs`")
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
