package fetching

import (
	"context"

	"github.com/elastic/beats/v7/libbeat/common"
)

// Factory can create fetcher instances based on configuration
type Factory interface {
	Create(*common.Config) (Fetcher, error)
}

// Fetcher represents a data fetcher.
type Fetcher interface {
	Fetch(context.Context) ([]Resource, error)
	Stop()
}

type Condition interface {
	Condition() bool
	Name() string
}

type Resource interface {
	GetID() (string, error)
	GetData() interface{}
}

type FetcherResult struct {
	Type string `json:"type"`
	// Golang 1.18 will introduce generics which will be useful for typing the resource field
	Resource interface{} `json:"resource"`
}

type ResourceMap map[string][]Resource

type BaseFetcherConfig struct {
	Name string `config:"name"`
}
