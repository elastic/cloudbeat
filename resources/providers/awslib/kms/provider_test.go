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

package kms

import (
	"context"
	"errors"
	"fmt"
	"testing"

	kmsClient "github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProviderTestSuite struct {
	suite.Suite

	log *logp.Logger
}
type mocks [2][]any
type kmsClientMockReturnVals map[string][]mocks

func TestProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)
	s.log = logp.NewLogger("cloudbeat_kms_provider_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {}

func (s *ProviderTestSuite) TearDownTest() {}

var keyId = "21c0ba99-3a6c-4f72-8ef8-8118d4804710"

func (s *ProviderTestSuite) TestProvider_DescribeBuckets() {
	var tests = []struct {
		name                    string
		regions                 []string
		kmsClientMockReturnVals kmsClientMockReturnVals
		expected                []awslib.AwsResource
		expectError             bool
	}{
		{
			name: "Should not return any resources when there aren't any keys",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{}}, nil}}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return a resource when key can't be described",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{
					{KeyId: &keyId},
				}}, nil}}},
				"DescribeKey": {{{mock.Anything, mock.Anything}, {nil, errors.New("some error")}}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return a resource when key is not symmetric",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{
					{KeyId: &keyId},
				}}, nil}}},
				// TODO: mock.matchBy({keyId}) ?
				"DescribeKey": {{{mock.Anything, mock.Anything}, {&kmsClient.DescribeKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: &keyId, KeySpec: "some string"}}, nil}}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return a resource when key rotation status can't be described",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{
					{KeyId: &keyId},
				}}, nil}}},
				"DescribeKey":          {{{mock.Anything, mock.Anything}, {&kmsClient.DescribeKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: &keyId, KeySpec: types.KeySpecSymmetricDefault}}, nil}}},
				"GetKeyRotationStatus": {{{mock.Anything, mock.Anything}, {nil, errors.New("some error")}}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			// TODO: add regions
			// need to change Mock.On to take multiple ListKeys calls
			name: "Should return a resource for a symmetric key",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{
					{KeyId: &keyId},
				}}, nil}}},
				"DescribeKey":          {{{mock.Anything, mock.Anything}, {&kmsClient.DescribeKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: &keyId, KeySpec: types.KeySpecSymmetricDefault}}, nil}}},
				"GetKeyRotationStatus": {{{mock.Anything, mock.Anything}, {&kmsClient.GetKeyRotationStatusOutput{KeyRotationEnabled: true}, nil}}},
			},
			expected:    []awslib.AwsResource{KmsInfo{KeyMetadata: types.KeyMetadata{KeyId: &keyId, KeySpec: types.KeySpecSymmetricDefault}, KeyRotationEnabled: true}},
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
	}

	for _, test := range tests {
		kmsClientMock := &MockClient{}
		for funcName, returnVals := range test.kmsClientMockReturnVals {
			for _, vals := range returnVals {
				kmsClientMock.On(funcName, vals[0]...).Return(vals[1]...).Once()
			}
		}

		kmsProvider := Provider{
			log:     s.log,
			clients: testhelper.CreateMockClients[Client](kmsClientMock, test.regions),
		}

		ctx := context.Background()

		results, err := kmsProvider.DescribeSymmetricKeys(ctx)
		if test.expectError {
			s.Error(err)
		} else {
			s.NoError(err)
		}

		// Using `ElementsMatch` instead of the usual `Equals` since iterating over the regions map does not produce a
		//	guaranteed order
		s.ElementsMatch(test.expected, results, fmt.Sprintf("Test '%s' failed, elements do not match", test.name))
	}
}
