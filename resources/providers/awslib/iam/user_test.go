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

package iam

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mocksReturnVals map[string][][]any

type expectedResult struct {
	Arn               string
	HasUsed           bool
	IsVirtualMFA      bool
	MfaActive         bool
	hasAttachedPolicy bool
}

func Test_GetUsers(t *testing.T) {
	tests := []struct {
		name            string
		mocksReturnVals mocksReturnVals
		expected        []expectedResult
		wantErr         bool
		expectedUsers   int
	}{
		{
			name: "Test 1: Should not return users",
			mocksReturnVals: mocksReturnVals{
				"ListUsers":           {{&iamsdk.ListUsersOutput{}, nil}},
				"GetCredentialReport": {{nil, nil}},
			},
			expected: []expectedResult{
				{
					Arn:               "",
					HasUsed:           false,
					IsVirtualMFA:      false,
					MfaActive:         true,
					hasAttachedPolicy: false,
				},
			},
			wantErr:       false,
			expectedUsers: 0,
		},
		{
			name: "Test 2: Should return 3 users",
			mocksReturnVals: mocksReturnVals{
				"ListUsers": {{&iamsdk.ListUsersOutput{Users: apiUsers}, nil}},
				"ListMFADevices": {
					{&iamsdk.ListMFADevicesOutput{MFADevices: virtualMfaDevices}, nil},
					{&iamsdk.ListMFADevicesOutput{MFADevices: mfaDevices}, nil},
					{nil, errors.New("no such user - root account")},
				},
				"GenerateCredentialReport": {{&iamsdk.GenerateCredentialReportOutput{}, nil}},
				"GetCredentialReport": {
					{nil, &smithy.GenericAPIError{Code: "ReportNotPresent"}},
					{CredentialsReportOutput, nil}},
				"ListUserPolicies": {
					{&iamsdk.ListUserPoliciesOutput{PolicyNames: []string{"inline-test-policy"}}, nil},
					{&iamsdk.ListUserPoliciesOutput{PolicyNames: []string{"inline-test-policy"}}, nil},
					{nil, errors.New("no such user - root account")},
				},
				"ListAttachedUserPolicies": {
					{&iamsdk.ListAttachedUserPoliciesOutput{AttachedPolicies: []types.AttachedPolicy{{PolicyName: aws.String("policy-name-1")}}}, nil},
					{&iamsdk.ListAttachedUserPoliciesOutput{AttachedPolicies: []types.AttachedPolicy{{PolicyName: aws.String("policy-name-2")}}}, nil},
					{nil, errors.New("no such user - root account")},
				},
				"GetUserPolicy": {
					{&iamsdk.GetUserPolicyOutput{PolicyDocument: aws.String("aws-test-policy"), PolicyName: aws.String("policy-name-1")}, nil},
					{&iamsdk.GetUserPolicyOutput{PolicyDocument: aws.String("aws-test-policy"), PolicyName: aws.String("policy-name-2")}, nil},
					{nil, errors.New("no such user - root account")},
				},
			},
			expected: []expectedResult{
				{
					Arn:               "arn:aws:iam::123456789012:user/user1",
					HasUsed:           true,
					IsVirtualMFA:      true,
					MfaActive:         true,
					hasAttachedPolicy: true,
				},
				{
					HasUsed:           true,
					MfaActive:         true,
					IsVirtualMFA:      false,
					Arn:               "arn:aws:iam::123456789012:user/user2",
					hasAttachedPolicy: true,
				},
				{
					HasUsed:           true,
					MfaActive:         false,
					IsVirtualMFA:      false,
					Arn:               "arn:aws:iam::1234567890:root",
					hasAttachedPolicy: false,
				},
			},
			wantErr:       false,
			expectedUsers: 3,
		},
		{
			name: "Test 3: Error should not return users",
			mocksReturnVals: mocksReturnVals{
				"ListUsers": {{&iamsdk.ListUsersOutput{}, errors.New("fail to list iam users")}},
			},
			expectedUsers: 0,
			expected:      []expectedResult{},
			wantErr:       true,
		},
		{
			name: "Test 4: Error failed to generate credentials report",
			mocksReturnVals: mocksReturnVals{
				"ListUsers": {{&iamsdk.ListUsersOutput{Users: apiUsers}, nil}},
				"GetCredentialReport": {{nil, &types.ServiceFailureException{
					Message: aws.String("service err"),
				}}},
			},
			expected:      []expectedResult{},
			wantErr:       true,
			expectedUsers: 0,
		},
	}

	for _, test := range tests {
		p := createProviderFromMockValues(test.mocksReturnVals)

		users, err := p.GetUsers(context.TODO())

		if !test.wantErr {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}

		assert.Equal(t, test.expectedUsers, len(users))
		for i, user := range users {
			assert.Equal(t, test.expected[i].Arn, user.(User).Arn)
			assert.Equal(t, test.expected[i].HasUsed, user.(User).AccessKeys[0].HasUsed)
			assert.Equal(t, test.expected[i].MfaActive, user.(User).MfaActive)
			assert.Equal(t, test.expected[i].hasAttachedPolicy, len(user.(User).AttachedPolicies)+len(user.(User).InlinePolicies) > 0)

			if test.expected[i].MfaActive {
				assert.Equal(t, test.expected[i].IsVirtualMFA, user.(User).MFADevices[0].IsVirtual)
			}
		}
	}
}

func createProviderFromMockValues(mockReturnValues mocksReturnVals) *Provider {
	mockedClient := MockClient{}
	for funcName, returnValues := range mockReturnValues {
		for _, values := range returnValues {
			mockedClient.On(funcName, mock.Anything, mock.Anything).Return(values...).Once()
		}
	}
	return &Provider{
		log:    logp.NewLogger("iam-provider"),
		client: &mockedClient,
	}
}
