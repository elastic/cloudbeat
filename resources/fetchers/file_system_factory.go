package fetchers

import (
	"encoding/gob"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

const (
	FileSystemType = "file-system"
)

func init() {
	manager.Factories.ListFetcherFactory(FileSystemType, &FileSystemFactory{})
	gob.Register(FileSystemResource{})
}

type FileSystemFactory struct {
}

func (f *FileSystemFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := FileFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *FileSystemFactory) CreateFrom(cfg FileFetcherConfig) (fetching.Fetcher, error) {
	fe := &FileSystemFetcher{
		cfg: cfg,
	}

	return fe, nil
}
