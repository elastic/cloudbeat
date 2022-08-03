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
	"context"
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"gopkg.in/yaml.v3"
	"time"
)

const DefaultNamespace = "default"

const ResultsDatastreamIndexPrefix = "logs-cloud_security_posture.findings"

const (
	InputTypeVanillaK8s = "cloudbeat/vanilla"
	InputTypeEks        = "cloudbeat/eks"
)

type Fetcher struct {
	Name string `config:"name"` // Name of the fetcher
}

type Fetchers struct {
	Vanilla []*config.C `config:"vanilla"` // Vanilla fetchers
	Eks     []*config.C `config:"eks"`     // EKS fetchers
}

type Config struct {
	Fetchers   Fetchers                `config:"fetchers"`
	KubeConfig string                  `config:"kube_config"`
	Period     time.Duration           `config:"period"`
	Processors processors.PluginConfig `config:"processors"`
	Streams    []Stream                `config:"streams"`
	Type       string                  `config:"type"`
}

type Stream struct {
	AWSConfig aws.ConfigAWS `config:",inline"`
	DataYaml  *DataYaml     `config:"data_yaml" yaml:"data_yaml" json:"data_yaml"`
}

type DataYaml struct {
	ActivatedRules *Benchmarks `config:"activated_rules" yaml:"activated_rules" json:"activated_rules"`
}

type Benchmarks struct {
	CisK8s []string `config:"cis_k8s,omitempty" yaml:"cis_k8s,omitempty" json:"cis_k8s,omitempty"`
	CisEks []string `config:"cis_eks,omitempty" yaml:"cis_eks,omitempty" json:"cis_eks,omitempty"`
}

var DefaultConfig = Config{
	Period: 4 * time.Hour,
}

func New(cfg *config.C) (Config, error) {
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
func (c *Config) Update(log *logp.Logger, cfg *config.C) error {
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

// GetActivatedRules returns the activated rules from the config. The activated rules are in yaml format.
func (c *Config) GetActivatedRules() (string, error) {
	dataYaml, err := yaml.Marshal(c.Streams[0].DataYaml)
	if err != nil {
		return "", err
	}

	return string(dataYaml), nil
}

// Datastream function to generate the datastream value
func Datastream(namespace string, indexPrefix string) string {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	return indexPrefix + "-" + namespace
}

type AwsConfigProvider interface {
	InitializeAWSConfig(ctx context.Context, cfg aws.ConfigAWS) (awssdk.Config, error)
}
