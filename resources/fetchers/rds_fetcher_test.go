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
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/rds"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
	"testing"
)

type RdsFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

type rdsMocksReturnVals map[string][]any

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
	var tests = []struct {
		name               string
		rdsMocksReturnVals rdsMocksReturnVals
		numExpectedResults int
	}{
		{
			name: "Should not get any DB instances",
			rdsMocksReturnVals: rdsMocksReturnVals{
				"DescribeDBInstances": {nil, errors.New("bad, very bad")},
			},
			numExpectedResults: 0,
		},
		{
			name: "Should get an Rds bucket",
			rdsMocksReturnVals: rdsMocksReturnVals{
				"DescribeDBInstances": {[]awslib.AwsResource{rds.DBInstance{Identifier: "id", Arn: "arn", StorageEncrypted: true, AutoMinorVersionUpgrade: true}}, nil},
			},
			numExpectedResults: 1,
		},
	}

	for _, test := range tests {
		rdsFetcherCfg := RdsFetcherConfig{
			AwsBaseFetcherConfig: fetching.AwsBaseFetcherConfig{},
		}

		rdsProviderMock := &rds.MockRds{}
		for funcName, returnVals := range test.rdsMocksReturnVals {
			rdsProviderMock.On(funcName, context.TODO()).Return(returnVals...)
		}

		rdsFetcher := RdsFetcher{
			log:        s.log,
			cfg:        rdsFetcherCfg,
			rds:        rdsProviderMock,
			resourceCh: s.resourceCh,
		}

		ctx := context.Background()

		err := rdsFetcher.Fetch(ctx, fetching.CycleMetadata{})
		s.NoError(err)

		results := testhelper.CollectResources(s.resourceCh)
		s.Equal(test.numExpectedResults, len(results))
	}
}
