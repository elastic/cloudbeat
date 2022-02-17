package fetchers

import (
	"context"
	"github.com/elastic/beats/v7/x-pack/osquerybeat/ext/osquery-extension/pkg/proc"
)

// Fetcher represents a data fetcher.
type Fetcher interface {
	Fetch(context.Context) ([]PolicyResource, error)
	Stop()
}

type FetcherCondition interface {
	Condition() bool
	Name() string
}

type PolicyResource interface {
	GetID() string
	GetData() interface{}
}

type FetcherResult struct {
	Type     string      `json:"type"`
	Resource interface{} `json:"resource"`
}

type ResourceMap map[string][]PolicyResource

type FileSystemResource struct {
	FileName string `json:"filename"`
	FileMode string `json:"mode"`
	Gid      string `json:"gid"`
	Uid      string `json:"uid"`
	Path     string `json:"path"`
	Inode    string `json:"inode"`
}

type ProcessResource struct {
	PID  string        `json:"pid"`
	Cmd  string        `json:"command"`
	Stat proc.ProcStat `json:"stat"`
}

type BaseFetcherConfig struct {
	Fetcher string `config:"fetcher"`
}
