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
	"os"
	"path/filepath"
	"time"

	cb_errors "github.com/elastic/cloudbeat/errors"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-libs/config"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
)

const DefaultNamespace = "default"

const ResultsDatastreamIndexPrefix = "logs-cloud_security_posture.findings"

var ErrBenchmarkNotSupported = cb_errors.New("benchmark is not supported")

type Fetcher struct {
	Name string `config:"name"` // Name of the fetcher
}

type Config struct {
	AWSConfig  aws.ConfigAWS           `config:",inline"`
	RuntimeCfg *RuntimeConfig          `config:"runtime_cfg"`
	Fetchers   []*config.C             `config:"fetchers"`
	KubeConfig string                  `config:"kube_config"`
	Period     time.Duration           `config:"period"`
	Processors processors.PluginConfig `config:"processors"`
	BundlePath string                  `config:"bundle_path"`
	Benchmark  string                  `config:"config.v1.benchmark"`
}

type RuntimeConfig struct {
	ActivatedRules *Benchmarks `config:"activated_rules" yaml:"activated_rules" json:"activated_rules"`
}

type Benchmarks struct {
	CisK8s []string `config:"cis_k8s,omitempty" yaml:"cis_k8s,omitempty" json:"cis_k8s,omitempty"`
	CisEks []string `config:"cis_eks,omitempty" yaml:"cis_eks,omitempty" json:"cis_eks,omitempty"`
	CisAws []string `config:"cis_aws,omitempty" yaml:"cis_aws,omitempty" json:"cis_aws,omitempty"`
}

func New(cfg *config.C) (*Config, error) {
	c, err := defaultConfig()
	if err != nil {
		return nil, err
	}

	if err := cfg.Unpack(&c); err != nil {
		return nil, err
	}

	if c.Benchmark != "" {
		if !isSupportedBenchmark(c.Benchmark) {
			return c, ErrBenchmarkNotSupported
		}
	}
	return c, nil
}

func defaultConfig() (*Config, error) {
	ret := &Config{
		Period:    4 * time.Hour,
		Benchmark: CIS_K8S,
	}

	bundle, err := getBundlePath()
	if err != nil {
		return nil, err
	}

	ret.BundlePath = bundle
	return ret, nil
}

func getBundlePath() (string, error) {
	// The bundle resides on the same location as the executable
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(ex), "bundle.tar.gz"), nil
}

// Datastream function to generate the datastream value
func Datastream(namespace string, indexPrefix string) string {
	if namespace == "" {
		namespace = DefaultNamespace
	}
	return indexPrefix + "-" + namespace
}

type AwsConfigProvider interface {
	InitializeAWSConfig(ctx context.Context, cfg aws.ConfigAWS, log *logp.Logger) (awssdk.Config, error)
}

func isSupportedBenchmark(benchmark string) bool {
	for _, s := range SupportedCIS {
		if benchmark == s {
			return true
		}
	}
	return false
}
