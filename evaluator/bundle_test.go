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

package evaluator

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/elastic/cloudbeat/config"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type BundleTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestBundleTestSuite(t *testing.T) {
	s := new(BundleTestSuite)
	s.log = logp.NewLogger("cloudbeat_bundle_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *BundleTestSuite) TestCreateServer() {
	_, err := StartServer(context.Background(), config.DefaultConfig)
	s.NoError(err)

	var tests = []struct {
		path               string
		expectedStatusCode string
	}{
		{
			"/bundles/bundle.tar.gz", "200 OK",
		},
		{
			"/bundles/notExistBundle.tar.gz", "404 Not Found",
		},
		{
			"/bundles/notExistBundle", "404 Not Found",
		},
	}

	time.Sleep(time.Second * 2)
	for _, test := range tests {
		target := ServerAddress + test.path
		client := &http.Client{}
		res, err := client.Get(target)

		s.NoError(err)
		s.Equal(test.expectedStatusCode, res.Status)
	}
}

func (s *BundleTestSuite) TestCreateServerWithRuntimeConfig() {
	invalidStreams := agentconfig.MustNewConfigFrom(`
    not_streams:
      - not_data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
              - c
              - d
              - e
`)

	configNoStreams, err := config.New(invalidStreams)
	s.NoError(err)

	invalidRuntimeConfig := agentconfig.MustNewConfigFrom(`
    streams:
      - data_yaml
`)
	configNoRuntimeConfig, err := config.New(invalidRuntimeConfig)
	s.Error(err)

	validStreams := agentconfig.MustNewConfigFrom(`
    streams:
      - data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
              - c
              - d
              - e
`)
	configWithRuntimeConfig, err := config.New(validStreams)
	s.NoError(err)

	var tests = []struct {
		name               string
		path               string
		expectedStatusCode string
		cfg                config.Config
	}{
		{
			"config missing data yaml", "/bundles/bundle.tar.gz", "200 OK", configNoRuntimeConfig,
		},
		{
			"config missing streams", "/bundles/bundle.tar.gz", "200 OK", configNoStreams,
		},
		{
			"valid config from string", "/bundles/bundle.tar.gz", "200 OK", configWithRuntimeConfig,
		},
		{
			"valid config struct", "/bundles/bundle.tar.gz", "200 OK",
			config.Config{
				Type: config.InputTypeVanillaK8s,
				Streams: []config.Stream{
					{
						RuntimeCfg: &config.RuntimeConfig{
							ActivatedRules: &config.Benchmarks{
								CisK8s: []string{
									"cis_1_1_1",
								},
							},
						},
					},
				},
			},
		},
		{
			"valid config struct", "/bundles/bundle.tar.gz", "200 OK",
			config.Config{
				Type: config.InputTypeEks,
				Streams: []config.Stream{
					{
						RuntimeCfg: &config.RuntimeConfig{
							ActivatedRules: &config.Benchmarks{
								CisEks: []string{
									"cis_1_1_1",
								},
							},
						},
					},
				},
			},
		},
	}

	time.Sleep(time.Second * 2)
	for _, test := range tests {
		s.Run(test.name, func() {
			server, err := StartServer(context.Background(), test.cfg)
			s.NoError(err)

			target := ServerAddress + test.path
			client := &http.Client{}
			res, err := client.Get(target)
			s.NoError(err)
			s.Equal(test.expectedStatusCode, res.Status)

			err = server.Shutdown(context.Background())
			s.NoError(err)
			time.Sleep(100 * time.Millisecond)
		})
	}
}

// TestCreateServerWithRuntimeConfig tests the creation of a server with a valid config
func (s *BundleTestSuite) TestCreateServerWithFetchersConfig() {
	validStreamsVanilla := agentconfig.MustNewConfigFrom(`
    type: cloudbeat/vanilla
    streams:
      - data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
`)
	configWithVanillaType, err := config.New(validStreamsVanilla)
	s.NoError(err)

	validStreamsEks := agentconfig.MustNewConfigFrom(`
    type: cloudbeat/eks
    streams:
      - data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
`)
	configWithEksType, err := config.New(validStreamsEks)
	s.NoError(err)

	var tests = []struct {
		name               string
		path               string
		expectedStatusCode string
		cfg                config.Config
	}{
		{
			"valid config struct (Vanilla)", "/bundles/bundle.tar.gz", "200 OK", configWithVanillaType,
		},
		{
			"valid config struct (EKS)", "/bundles/bundle.tar.gz", "200 OK", configWithEksType,
		},
		{
			"valid config struct", "/bundles/bundle.tar.gz", "200 OK",
			config.Config{
				Type: config.InputTypeVanillaK8s,
				Streams: []config.Stream{
					{
						RuntimeCfg: &config.RuntimeConfig{
							ActivatedRules: &config.Benchmarks{
								CisK8s: []string{
									"cis_1_1_1",
								},
							},
						},
					},
				},
			},
		},
		{
			"valid config struct", "/bundles/bundle.tar.gz", "200 OK",
			config.Config{
				Type: config.InputTypeEks,
				Streams: []config.Stream{
					{
						RuntimeCfg: &config.RuntimeConfig{
							ActivatedRules: &config.Benchmarks{
								CisEks: []string{
									"cis_1_1_1",
								},
							},
						},
					},
				},
			},
		},
	}

	time.Sleep(time.Second * 2)
	for _, test := range tests {
		s.Run(test.name, func() {
			server, err := StartServer(context.Background(), test.cfg)
			s.NoError(err)

			target := ServerAddress + test.path
			client := &http.Client{}
			res, err := client.Get(target)
			s.NoError(err)
			s.Equal(test.expectedStatusCode, res.Status)

			err = server.Shutdown(context.Background())
			s.NoError(err)
			time.Sleep(100 * time.Millisecond)
		})
	}
}
