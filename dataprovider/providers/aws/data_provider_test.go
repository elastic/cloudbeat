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

package aws

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/cloudbeat/dataprovider/types"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	accountName = "accountName"
	accountId   = "accountId"
)

type AwsDataProviderTestSuite struct {
	suite.Suite
	log *logp.Logger
}

func TestAwsDataProviderTestSuite(t *testing.T) {
	s := new(AwsDataProviderTestSuite)
	s.log = logp.NewLogger("cloudbeat_aws_data_provider_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *AwsDataProviderTestSuite) SetupTest() {}

func (s *AwsDataProviderTestSuite) TearDownTest() {}

func (s *AwsDataProviderTestSuite) TestAwsDataProvider_FetchData() {
	tests := []struct {
		name        string
		options     []Option
		resource    string
		id          string
		expected    types.Data
		expectError bool
	}{
		{
			name: "get data",
			options: []Option{
				WithLogger(s.log),
				WithAccount(accountName, accountId),
			},
			expected: types.Data{
				ResourceID: "",
				VersionInfo: version.CloudbeatVersionInfo{
					Version: version.CloudbeatVersion(),
					Policy:  version.PolicyVersion(),
				},
			},
		},
	}

	for _, test := range tests {
		p := New(test.options...)
		result, err := p.FetchData(test.resource, test.id)
		if test.expectError {
			s.Error(err)
			return
		}
		s.NoError(err)
		s.Equal(result, test.expected)

	}
}

func TestAWSDataProvider_EnrichEvent(t *testing.T) {
	options := []Option{
		WithLogger(logp.NewLogger("")),
		WithAccount(accountName, accountId),
	}

	k := New(options...)
	e := &beat.Event{
		Fields: mapstr.M{},
	}
	err := k.EnrichEvent(e)
	assert.NoError(t, err)

	accountID, err := e.Fields.GetValue(cloudAccountIdField)
	assert.NoError(t, err)
	assert.Equal(t, "accountId", accountID)

	accountName, err := e.Fields.GetValue(cloudAccountNameField)
	assert.NoError(t, err)
	assert.Equal(t, "accountName", accountName)

	cloud, err := e.Fields.GetValue(cloudProviderField)
	assert.NoError(t, err)
	assert.Equal(t, "aws", cloud)
}
