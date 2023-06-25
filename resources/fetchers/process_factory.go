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
	"os"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	ProcessType = "process"
)

type ProcessFactory struct {
}

func init() {
	fetchersManager.Factories.RegisterFactory(ProcessType, &ProcessFactory{})
}

func (f *ProcessFactory) Create(log *logp.Logger, c *config.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting ProcessFactory.Create")

	cfg := ProcessFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	log.Infof("Process Fetcher created with the following config:"+
		"\n Name: %s\nDirectory: %s\nRequiredProcesses: %s", cfg.Name, cfg.Directory, cfg.RequiredProcesses)
	return f.CreateFrom(log, cfg, ch)
}

func (f *ProcessFactory) CreateFrom(log *logp.Logger, cfg ProcessFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	fe := &ProcessesFetcher{
		log:        log,
		cfg:        cfg,
		Fs:         os.DirFS(cfg.Directory),
		resourceCh: ch,
	}

	return fe, nil
}
