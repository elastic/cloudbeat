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

package beater

import (
	"testing"

	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type ValidatorTestSuite struct {
	suite.Suite

	log *logp.Logger
	sut *validator
}

func TestValidatorTestSuite(t *testing.T) {
	s := new(ValidatorTestSuite)
	s.log = logp.NewLogger("cloudbeat_validator_test_suite")
	s.sut = &validator{}

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ValidatorTestSuite) TestConfig() {
	configWithBenchmark := config.MustNewConfigFrom(`
config:
  v1:
    benchmark: cis_k8s
`)

	testcases := []struct {
		err bool
		cfg *config.C
	}{
		{
			false,
			config.NewConfig(),
		},
		{
			false,
			configWithBenchmark,
		},
	}

	for _, tcase := range testcases {
		err := s.sut.Validate(tcase.cfg)
		s.Equal(tcase.err, err != nil)
	}
}
