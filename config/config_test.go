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

package config

import (
	"testing"
	"time"

	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestConfigTestSuite(t *testing.T) {
	s := new(ConfigTestSuite)
	s.log = logp.NewLogger("cloudbeat_config_test_suite")
	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ConfigTestSuite) SetupTest() {
}

func (s *ConfigTestSuite) TestNew() {
	var tests = []struct {
		config                 string
		expectedActivatedRules []string
		expectedType           string
		expectedAccessKey      string
		expectedSecret         string
		expectedSessionToken   string
	}{
		{
			`
   type : cloudbeat/cis_k8s
   streams:
    - data_yaml:
        activated_rules:
          cis_k8s:
            - a
            - b
            - c
            - d
            - e
      access_key_id: key
      secret_access_key: secret
      session_token: session
`,
			[]string{"a", "b", "c", "d", "e"},
			"cloudbeat/cis_k8s",
			"key",
			"secret",
			"session",
		},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expectedType, c.Type)
		s.Equal(test.expectedActivatedRules, c.Streams[0].RuntimeCfg.ActivatedRules.CisK8s)
		s.Equal(test.expectedAccessKey, c.Streams[0].AWSConfig.AccessKeyID)
		s.Equal(test.expectedSecret, c.Streams[0].AWSConfig.SecretAccessKey)
		s.Equal(test.expectedSessionToken, c.Streams[0].AWSConfig.SessionToken)
	}
}

func (s *ConfigTestSuite) TestDataYamlExists() {
	var tests = []struct {
		config   string
		expected bool
	}{
		{
			`
  streams:
    - data_yaml:
        activated_rules:
          cis_k8s:
            - a
            - b
            - c
            - d
            - e
`,
			true,
		},
		{
			`
  streams:
    - not_data_yaml:
        something: true
`,
			false,
		},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expected, c.Streams[0].RuntimeCfg != nil)
	}
}

func (s *ConfigTestSuite) TestConfigUpdate() {
	configYml := `
    streams:
      - data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
              - c
              - d
              - e
        access_key_id: old_key
        secret_access_key: old_secret
        session_token: old_session`

	var tests = []struct {
		update                 string
		expectedActivatedRules []string
		expectedAccessKey      string
		expectedSecret         string
		expectedSessionToken   string
	}{
		{
			`
          streams:
            - data_yaml:
                activated_rules:
                  cis_k8s:
                    - a	
                    - b
                    - c
                    - d
              access_key_id: new_key
              secret_access_key: new_secret
              session_token: new_session
        `,
			[]string{"a", "b", "c", "d"},
			"new_key",
			"new_secret",
			"new_session",
		},
		{
			`
          streams:
            - data_yaml:
                activated_rules:
                  cis_k8s:
                    - b	
                    - c
                    - d
                    - e
                    - f
              access_key_id: new_key1
              secret_access_key: new_secret1
              session_token: new_session1
        `,
			[]string{"b", "c", "d", "e", "f"},
			"new_key1",
			"new_secret1",
			"new_session1",
		},
		{
			`
          streams:
            - data_yaml:
                activated_rules:
                  cis_k8s: []
              access_key_id: 
              secret_access_key:
              session_token:
        `,
			[]string{},
			"",
			"",
			"",
		},
		//We're not currently aware of a scenario where we receive
		//multiple input streams, but still to make sure that it doesn't
		//break for this scenario.
		{
			`
          streams:
            - data_yaml:
                activated_rules:
                  cis_k8s:
                    - x	
                    -  "y" # Just YAML 1.1 things
                    - z
              access_key_id: first_stream_key
              secret_access_key: first_stream_secret
              session_token: first_stream_session  
            - data_yaml:
                activated_rules:
                  cis_k8s:
                    - f	
                    - g
                    - h
                    - i
                    - j
              access_key_id: second_stream_key
              secret_access_key: second_stream_secret
              session_token: second_stream_session
        `,
			[]string{"x", "y", "z"},
			"first_stream_key",
			"first_stream_secret",
			"first_stream_session",
		},
	}

	cfg, err := config.NewConfigFrom(configYml)
	s.NoError(err)

	c, err := New(cfg)
	s.NoError(err)

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.update)
		s.NoError(err)

		err = c.Update(s.log, cfg)
		s.NoError(err)

		s.Equal(test.expectedActivatedRules, c.Streams[0].RuntimeCfg.ActivatedRules.CisK8s)
		s.Equal(test.expectedAccessKey, c.Streams[0].AWSConfig.AccessKeyID)
		s.Equal(test.expectedSecret, c.Streams[0].AWSConfig.SecretAccessKey)
		s.Equal(test.expectedSessionToken, c.Streams[0].AWSConfig.SessionToken)
	}
}

// TestConfigUpdateIsolated tests whether updates made to a config from
// are isolated; only those parts of the config specified in the incoming
// config should get updated.
func (s *ConfigTestSuite) TestConfigUpdateIsolated() {
	configYml := `
    period: 10s
    kube_config: some_path
    streams:
      - data_yaml:
          activated_rules:
            cis_k8s:
              - a
              - b
              - c
              - d
              - e
        access_key_id: old_key
        secret_access_key: old_secret
        session_token: old_session`

	var tests = []struct {
		update               string
		expectedPeriod       time.Duration
		expectedKubeConfig   string
		expectedCISK8S       []string
		expectedAccessKey    string
		expectedSecret       string
		expectedSessionToken string
	}{
		{
			update: `
           streams:
             - data_yaml:
                 activated_rules:
                   cis_k8s:
                     - a
                     - b
                     - c
                     - d
               access_key_id: new_key
               secret_access_key: new_secret
               session_token: new_session`,
			expectedPeriod:       10 * time.Second,
			expectedKubeConfig:   "some_path",
			expectedCISK8S:       []string{"a", "b", "c", "d"},
			expectedAccessKey:    "new_key",
			expectedSecret:       "new_secret",
			expectedSessionToken: "new_session",
		},
		{
			update:               `period: 4h`,
			expectedPeriod:       4 * time.Hour,
			expectedKubeConfig:   "some_path",
			expectedCISK8S:       []string{"a", "b", "c", "d"},
			expectedAccessKey:    "new_key",
			expectedSecret:       "new_secret",
			expectedSessionToken: "new_session",
		},
		{
			update: `
           kube_config: some_other_path
           streams:
             - data_yaml:
                 activated_rules:
                   cis_k8s:
                     - a
                     - b
                     - c
                     - d
                     - e`,
			expectedPeriod:       4 * time.Hour,
			expectedKubeConfig:   "some_other_path",
			expectedCISK8S:       []string{"a", "b", "c", "d", "e"},
			expectedAccessKey:    "",
			expectedSecret:       "",
			expectedSessionToken: "",
		},
	}

	cfg, err := config.NewConfigFrom(configYml)
	s.NoError(err)

	c, err := New(cfg)
	s.NoError(err)

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.update)
		s.NoError(err)

		err = c.Update(s.log, cfg)
		s.NoError(err)

		s.Equal(test.expectedPeriod, c.Period)
		s.Equal(test.expectedKubeConfig, c.KubeConfig)
		s.Equal(test.expectedCISK8S, c.Streams[0].RuntimeCfg.ActivatedRules.CisK8s)
		s.Equal(test.expectedAccessKey, c.Streams[0].AWSConfig.AccessKeyID)
		s.Equal(test.expectedSecret, c.Streams[0].AWSConfig.SecretAccessKey)
		s.Equal(test.expectedSessionToken, c.Streams[0].AWSConfig.SessionToken)
	}
}

func (s *ConfigTestSuite) TestConfigDataYaml() {
	var tests = []struct {
		config   string
		expected []string
	}{
		{
			`
  streams:
    - data_yaml:
        activated_rules:
          cis_k8s:
            - a
            - b
            - c
            - d
`, []string{"a", "b", "c", "d"}},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		dy, err := c.GetActivatedRules()
		s.NoError(err)

		s.Equal(test.expected, dy.CisK8s)
	}
}

func (s *ConfigTestSuite) TestConfigPeriod() {
	var tests = []struct {
		config         string
		expectedPeriod time.Duration
	}{
		{"", 4 * time.Hour},
		{"period: 50s", 50 * time.Second},
		{"period: 5m", 5 * time.Minute},
		{"period: 2h", 2 * time.Hour},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expectedPeriod, c.Period)
	}
}

func (s *ConfigTestSuite) TestActivatedRulesFrameWork() {
	var tests = []struct {
		config                    string
		expectedActivatedRules    []string
		expectedEksActivatedRules []string
		expectedType              string
	}{
		{
			`
type: cloudbeat/cis_k8s
streams:
  - data_yaml:
      activated_rules:
        cis_k8s:
          - a
          - b
`,
			[]string{"a", "b"},
			nil,
			"cloudbeat/cis_k8s",
		},
		{
			`
type: cloudbeat/cis_eks
streams:
  - data_yaml:
      activated_rules:
        cis_eks:
          - a
          - b
`,
			nil,
			[]string{"a", "b"},
			"cloudbeat/cis_eks",
		},
	}

	for _, test := range tests {
		cfg, err := config.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expectedType, c.Type)
		s.Equal(test.expectedActivatedRules, c.Streams[0].RuntimeCfg.ActivatedRules.CisK8s)
		s.Equal(test.expectedEksActivatedRules, c.Streams[0].RuntimeCfg.ActivatedRules.CisEks)
	}
}
