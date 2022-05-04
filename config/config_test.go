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

//go:build !integration
// +build !integration

package config

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
}

func (s *ConfigTestSuite) TestNew() {
	var tests = []struct {
		config           string
		expectedPatterns []string
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
`,
			[]string{"a", "b", "c", "d"},
		},
	}

	for _, test := range tests {
		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		c, err := New(cfg)
		s.NoError(err)

		s.Equal(test.expectedPatterns, c.Streams[0].DataYaml.ActivatedRules.CISK8S)
	}
}
