package fetchers

import (
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
)

type KubeFactory struct {
}

func (f *KubeFactory) Create(c *common.Config) (Fetcher, error) {
	cfg := KubeApiFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *KubeFactory) CreateFrom(cfg KubeApiFetcherConfig) (Fetcher, error) {
	fe := &KubeFetcher{
		cfg:      cfg,
		watchers: make([]kubernetes.Watcher, 0),
	}

	return fe, nil
}
