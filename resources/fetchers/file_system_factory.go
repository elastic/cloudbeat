package fetchers

import (
	"github.com/elastic/beats/v7/libbeat/common"
)

type FileSystemFactory struct {
}

func (f *FileSystemFactory) Create(c common.Config) (Fetcher, error) {
	cfg := FileFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *FileSystemFactory) CreateFrom(cfg FileFetcherConfig) (Fetcher, error) {
	fe := &FileSystemFetcher{
		cfg: cfg,
	}

	return fe, nil
}
