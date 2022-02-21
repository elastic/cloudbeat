package fetchers

import (
	"context"
)

// Fetcher represents a data fetcher.
type Fetcher interface {
	Fetch(context.Context) ([]FetchedResource, error)
	Stop()
}

type FetcherCondition interface {
	Condition() bool
	Name() string
}

type FetchedResource interface {
	GetID() string
	GetData() interface{}
}

type FetcherResult struct {
	Type     string      `json:"type"`
	Resource interface{} `json:"resource"`
}

type ResourceMap map[string][]FetchedResource

type BaseFetcherConfig struct {
	Fetcher string `config:"fetcher"`
}
