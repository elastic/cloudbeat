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
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/processors"
	"gopkg.in/yaml.v3"
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
	DataYaml *struct {
		ActivatedRules struct {
			CISK8S []string `config:"cis_k8s" yaml:"cis_k8s" json:"cis_k8s"`
		} `config:"activated_rules" yaml:"activated_rules" json:"activated_rules"`
	} `config:"data_yaml" yaml:"data_yaml" json:"data_yaml"`
}

var DefaultConfig = Config{
	Period: 4 * time.Hour,
}

func New(cfg *common.Config) (Config, error) {
	c := DefaultConfig

	if err := cfg.Unpack(&c); err != nil {
		return c, err
	}

	return c, nil
}

// Update replaces values of those keys in the current config which are
// present in the incoming config.
//
// NOTE(yashtewari): This will be removed with the planned update to restart the
// beat with the new config.
func (c *Config) Update(log *logp.Logger, cfg *common.Config) error {
	log.Infof("Updating config with the following keys: %v", cfg.FlattenedKeys())

	if err := cfg.Unpack(&c); err != nil {
		return err
	}

	// Check if the incoming config has streams.
	if cfg.HasField("streams") {
		uc, err := New(cfg)
		if err != nil {
			return err
		}

		c.Streams = uc.Streams
	}

	return nil
}

func (c *Config) DataYaml() (string, error) {
	// TODO(yashtewari): Figure out the scenarios in which the integration sends
	// multiple input streams. Since only one instance of our integration is allowed per
	// agent policy, is it even possible that multiple input streams are received?
	y, err := yaml.Marshal(c.Streams[0].DataYaml)
	if err != nil {
		return "", err
	}

	return string(y), nil
}

// Datastream function to generate the datastream value
func Datastream(namespace string, indexPrefix string) string {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	return indexPrefix + "-" + namespace
}
