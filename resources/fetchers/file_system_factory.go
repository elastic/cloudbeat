// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package fetchers

import (
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
	common "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	FileSystemType = "file-system"
)

func init() {
	manager.Factories.ListFetcherFactory(FileSystemType, &FileSystemFactory{})
}

type FileSystemFactory struct {
}

func (f *FileSystemFactory) Create(c *common.C) (fetching.Fetcher, error) {
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

	logp.L().Infof("File-System Fetcher created with the following config:"+
		"\n Name: %s\nPatterns: %s", cfg.Name, cfg.Patterns)
	return fe, nil
}
