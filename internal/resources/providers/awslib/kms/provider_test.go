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
	"testing"

	kmsClient "github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type ProviderTestSuite struct {
	suite.Suite
}
type (
	mocks                   [2][]any
	kmsClientMockReturnVals map[string][]mocks
)

func TestProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)
	suite.Run(t, s)
}

func (s *ProviderTestSuite) SetupTest() {}

func (s *ProviderTestSuite) TearDownTest() {}

var (
	keyId1 = "21c0ba99-3a6c-4f72-8ef8-8118d4804710"
	keyId2 = "21c0ba99-3a6c-4f72-8ef8-8118d4804711"
)

func (s *ProviderTestSuite) TestProvider_DescribeSymmetricKeys() {
	tests := []struct {
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
			name: "Should not return any resources keys can't be listed",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {{{mock.Anything, mock.Anything}, {nil, errors.New("some error")}}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: true,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return a resource when key can't be described",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{
					{KeyId: &keyId1},
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
					{KeyId: &keyId1},
				}}, nil}}},
				"DescribeKey": {{{mock.Anything, mock.Anything}, {&kmsClient.DescribeKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: &keyId1, KeySpec: "some string"}}, nil}}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return a resource when key rotation status can't be described",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{
					{KeyId: &keyId1},
				}}, nil}}},
				"DescribeKey":          {{{mock.Anything, mock.Anything}, {&kmsClient.DescribeKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: &keyId1, KeySpec: types.KeySpecSymmetricDefault}}, nil}}},
				"GetKeyRotationStatus": {{{mock.Anything, mock.Anything}, {nil, errors.New("some error")}}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should not return a resource when key is managed by AWS",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{
					{KeyId: &keyId1},
				}}, nil}}},
				"DescribeKey":          {{{mock.Anything, mock.Anything}, {&kmsClient.DescribeKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: &keyId1, KeySpec: types.KeySpecSymmetricDefault, KeyManager: types.KeyManagerTypeAws}}, nil}}},
				"GetKeyRotationStatus": {{{mock.Anything, mock.Anything}, {nil, errors.New("some error")}}},
			},
			expected:    []awslib.AwsResource(nil),
			expectError: false,
			regions:     []string{awslib.DefaultRegion},
		},
		{
			name: "Should return a resource for a symmetric key",
			kmsClientMockReturnVals: kmsClientMockReturnVals{
				"ListKeys": {
					{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{
						{KeyId: &keyId1},
					}}, nil}},
					{{mock.Anything, mock.Anything}, {&kmsClient.ListKeysOutput{Keys: []types.KeyListEntry{
						{KeyId: &keyId2},
					}}, nil}},
				},
				"DescribeKey": {
					{{mock.Anything, mock.MatchedBy(MatchDescribeKeyInput(keyId1))}, {&kmsClient.DescribeKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: &keyId1, KeySpec: types.KeySpecSymmetricDefault, KeyManager: types.KeyManagerTypeCustomer}}, nil}},
					{{mock.Anything, mock.MatchedBy(MatchDescribeKeyInput(keyId2))}, {&kmsClient.DescribeKeyOutput{KeyMetadata: &types.KeyMetadata{KeyId: &keyId2, KeySpec: types.KeySpecSymmetricDefault, KeyManager: types.KeyManagerTypeCustomer}}, nil}},
				},
				"GetKeyRotationStatus": {
					{{mock.Anything, mock.MatchedBy(MatchGetKeyRotationStatusInput(keyId1))}, {&kmsClient.GetKeyRotationStatusOutput{KeyRotationEnabled: true}, nil}},
					{{mock.Anything, mock.MatchedBy(MatchGetKeyRotationStatusInput(keyId2))}, {&kmsClient.GetKeyRotationStatusOutput{KeyRotationEnabled: true}, nil}},
				},
			},
			expected: []awslib.AwsResource{
				KmsInfo{KeyMetadata: types.KeyMetadata{KeyId: &keyId1, KeySpec: types.KeySpecSymmetricDefault, KeyManager: types.KeyManagerTypeCustomer}, KeyRotationEnabled: true, region: "us-east-1"},
				KmsInfo{KeyMetadata: types.KeyMetadata{KeyId: &keyId2, KeySpec: types.KeySpecSymmetricDefault, KeyManager: types.KeyManagerTypeCustomer}, KeyRotationEnabled: true, region: "us-east-2"},
			},
			expectError: false,
			regions:     []string{"us-east-1", "us-east-2"},
		},
	}

	for _, test := range tests {
		mockClients := make(map[string]Client, len(test.regions))
		for i, region := range test.regions {
			kmsClientMock := &MockClient{}
			for funcName, returnVals := range test.kmsClientMockReturnVals {
				kmsClientMock.On(funcName, returnVals[i][0]...).Return(returnVals[i][1]...)
			}
			mockClients[region] = kmsClientMock
		}

		kmsProvider := Provider{
			log:     testhelper.NewLogger(s.T()),
			clients: mockClients,
		}

		ctx := context.Background()

		results, err := kmsProvider.DescribeSymmetricKeys(ctx)
		if test.expectError {
			s.Require().Error(err)
		} else {
			s.Require().NoError(err)
		}

		// Using `ElementsMatch` instead of the usual `Equals` since iterating over the regions map does not produce a
		//	guaranteed order
		s.ElementsMatch(test.expected, results, "Test '%s' failed, elements do not match", test.name)
	}
}

func MatchDescribeKeyInput(keyId string) func(k *kmsClient.DescribeKeyInput) bool {
	return func(k *kmsClient.DescribeKeyInput) bool {
		return *k.KeyId == keyId
	}
}

func MatchGetKeyRotationStatusInput(keyId string) func(k *kmsClient.GetKeyRotationStatusInput) bool {
	return func(k *kmsClient.GetKeyRotationStatusInput) bool {
		return *k.KeyId == keyId
	}
}
