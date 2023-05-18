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
	"github.com/elastic/cloudbeat/resources/providers/awslib/kms"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
)

type KmsFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

type KmsMocksReturnVals map[string][]any

func TestKmsFetcherTestSuite(t *testing.T) {
	s := new(KmsFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_kms_fetcher_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *KmsFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *KmsFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *KmsFetcherTestSuite) TestFetcher_Fetch() {
	var tests = []struct {
		name               string
		kmsMocksReturnVals KmsMocksReturnVals
		numExpectedResults int
	}{
		{
			name: "Should not return keys from KMS",
			kmsMocksReturnVals: KmsMocksReturnVals{
				"DescribeSymmetricKeys": {nil, errors.New("some error")},
			},
			numExpectedResults: 0,
		},
		{
			name: "Should return a key from KMS",
			kmsMocksReturnVals: KmsMocksReturnVals{
				"DescribeSymmetricKeys": {[]awslib.AwsResource{kms.KmsInfo{}}, nil},
			},
			numExpectedResults: 1,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			kmsFetcherCfg := KmsFetcherConfig{
				AwsBaseFetcherConfig: fetching.AwsBaseFetcherConfig{},
			}

			kmsProviderMock := &kms.MockKMS{}
			for funcName, returnVals := range test.kmsMocksReturnVals {
				kmsProviderMock.On(funcName, context.TODO()).Return(returnVals...)
			}

			kmsFetcher := KmsFetcher{
				log:        s.log,
				cfg:        kmsFetcherCfg,
				kms:        kmsProviderMock,
				resourceCh: s.resourceCh,
			}

			ctx := context.Background()

			err := kmsFetcher.Fetch(ctx, fetching.CycleMetadata{})
			s.NoError(err)

			results := testhelper.CollectResources(s.resourceCh)
			s.Equal(test.numExpectedResults, len(results))
		})
	}
}
