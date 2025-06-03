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
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/kms"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type KmsFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

type KmsMocksReturnVals map[string][]any

func TestKmsFetcherTestSuite(t *testing.T) {
	s := new(KmsFetcherTestSuite)

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

	t := s.T()
	for _, test := range tests {
		s.Run(test.name, func() {
			ctx := t.Context()
			kmsProviderMock := &kms.MockKMS{}
			for funcName, returnVals := range test.kmsMocksReturnVals {
				kmsProviderMock.On(funcName, ctx).Return(returnVals...)
			}

			kmsFetcher := KmsFetcher{
				log:        testhelper.NewLogger(s.T()),
				kms:        kmsProviderMock,
				resourceCh: s.resourceCh,
			}

			err := kmsFetcher.Fetch(ctx, cycle.Metadata{})
			s.Require().NoError(err)

			results := testhelper.CollectResources(s.resourceCh)
			s.Len(results, test.numExpectedResults)
		})
	}
}

func (s *KmsFetcherTestSuite) TestKmsResource_GetMetadata() {
	validFrom := aws.Time(time.Date(1992, 9, 1, 0, 0, 0, 0, time.UTC))
	validTo := aws.Time(time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC))
	r := KmsResource{
		key: kms.KmsInfo{
			KeyMetadata: types.KeyMetadata{
				KeyId:        aws.String("test-key-id"),
				Arn:          aws.String("test-key-arn"),
				KeyUsage:     types.KeyUsageTypeEncryptDecrypt,
				KeySpec:      types.KeySpecEccNistP256,
				ValidTo:      validTo,
				CreationDate: validFrom,
			},
		},
	}
	meta, err := r.GetMetadata()
	s.Require().NoError(err)
	s.Equal(fetching.ResourceMetadata{ID: "test-key-arn", Type: "key-management", SubType: "aws-kms", Name: "test-key-id"}, meta)

	m, err := r.GetElasticCommonData()
	s.Require().NoError(err)
	s.Len(m, 4)
	s.Contains(m, "cloud.service.name")
	s.Equal(types.KeySpecEccNistP256, m["x509.public_key_algorithm"])
	s.Equal(validTo, m["x509.not_after"])
	s.Equal(validFrom, m["x509.not_before"])
}
