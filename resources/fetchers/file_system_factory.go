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
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/user"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	FileSystemType = "file-system"
)

func init() {
	fetchersManager.Factories.RegisterFactory(FileSystemType, &FileSystemFactory{})
}

type FileSystemFactory struct {
}

func (f *FileSystemFactory) Create(log *logp.Logger, c *config.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting FileSystemFactory.Create")

	cfg := FileFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(log, cfg, ch)
}

func (f *FileSystemFactory) CreateFrom(log *logp.Logger, cfg FileFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	fe := &FileSystemFetcher{
		log:        log,
		cfg:        cfg,
		resourceCh: ch,
		osUser:     user.NewOSUserUtil(),
	}

	log.Infof("File-System Fetcher created with the following config:"+
		"\n Name: %s\nPatterns: %s", cfg.Name, cfg.Patterns)
	return fe, nil
}
