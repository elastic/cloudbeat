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
	"context"
	"errors"
	"testing"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/rds"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RdsFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

type rdsMocksReturnVals map[string][]any

var (
	dbInstance1 = rds.DBInstance{Identifier: "id", Arn: "arn", StorageEncrypted: true, AutoMinorVersionUpgrade: true}
	dbInstance2 = rds.DBInstance{Identifier: "id2", Arn: "arn2", StorageEncrypted: false, AutoMinorVersionUpgrade: false}
)

func TestRdsFetcherTestSuite(t *testing.T) {
	s := new(RdsFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_rds_fetcher_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *RdsFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *RdsFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *RdsFetcherTestSuite) TestFetcher_Fetch() {
	tests := []struct {
		name               string
		rdsMocksReturnVals rdsMocksReturnVals
		expected           []fetching.ResourceInfo
	}{
		{
			name: "Should not get any DB instances",
			rdsMocksReturnVals: rdsMocksReturnVals{
				"DescribeDBInstances": {nil, errors.New("bad, very bad")},
			},
			expected: []fetching.ResourceInfo(nil),
		},
		{
			name: "Should get an RDS DB instance",
			rdsMocksReturnVals: rdsMocksReturnVals{
				"DescribeDBInstances": {[]awslib.AwsResource{dbInstance1, dbInstance2}, nil},
			},
			expected: []fetching.ResourceInfo{{Resource: RdsResource{dbInstance: dbInstance1}}, {Resource: RdsResource{dbInstance: dbInstance2}}},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			rdsFetcherCfg := RdsFetcherConfig{
				AwsBaseFetcherConfig: fetching.AwsBaseFetcherConfig{},
			}

			m := &rds.MockRds{}
			for fn, rdsMocksReturnVals := range test.rdsMocksReturnVals {
				m.On(fn, mock.Anything).Return(rdsMocksReturnVals...)
			}

			rdsFetcher := RdsFetcher{
				log:        s.log,
				cfg:        rdsFetcherCfg,
				resourceCh: s.resourceCh,
				provider:   m,
			}

			ctx := context.Background()

			err := rdsFetcher.Fetch(ctx, fetching.CycleMetadata{})
			s.NoError(err)

			results := testhelper.CollectResources(s.resourceCh)
			s.ElementsMatch(test.expected, results)
		})
	}
}
