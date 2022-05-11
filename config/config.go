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

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/processors"
)

const DefaultNamespace = "default"

const ResultsDatastreamIndexPrefix = "logs-cloud_security_posture.findings"

type Config struct {
	KubeConfig string                  `config:"kube_config"`
	Period     time.Duration           `config:"period"`
	Processors processors.PluginConfig `config:"processors"`
	Fetchers   []*common.Config        `config:"fetchers"`

	Streams []Stream `config:"streams"`
}

type Stream struct {
	DataYaml struct {
		ActivatedRules struct {
			CISK8S []string `config:"cis_k8s"`
		} `config:"activated_rules"`
	} `config:"data_yaml"`
}

var DefaultConfig = Config{
	Period: 10 * time.Second,
}

func New(cfg *common.Config) (Config, error) {
	c := DefaultConfig

	if err := cfg.Unpack(&c); err != nil {
		return c, err
	}

	return c, nil
}

func (c *Config) Update(cfg *common.Config) error {
	if err := cfg.Unpack(&c); err != nil {
		return err
	}

	return nil
}

// Datastream function to generate the datastream value
func Datastream(namespace string, indexPrefix string) string {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	return indexPrefix + "-" + namespace
}
