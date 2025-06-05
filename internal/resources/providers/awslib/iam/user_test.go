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
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

const credentialsReportContent = `user,arn,user_creation_time,password_enabled,password_last_used,password_last_changed,password_next_rotation,mfa_active,access_key_1_active,access_key_1_last_rotated,access_key_1_last_used_date,access_key_2_active,access_key_2_last_rotated,access_key_2_last_used_date,cert_1_active,cert_2_active\n
<root_account>,arn:aws:iam::1234567890:root,1970-01-01T00:00:00+00:00,true,2022-01-02T00:00:00+00:00,1970-01-01T00:00:00+00:00,2022-01-03T00:00:00+00:00,false,true,1970-01-01T00:00:00+00:00,2022-01-04T00:00:00+00:00,true,1970-01-01T00:00:00+00:00,2022-01-05T00:00:00+00:00,true,true\n
user1,arn:aws:iam::1234567890:user/user1,2022-01-01T00:00:00+00:00,true,2022-01-02T00:00:00+00:00,2022-01-03T00:00:00+00:00,2022-01-04T00:00:00+00:00,true,true,2022-01-05T00:00:00+00:00,2022-01-06T00:00:00+00:00,true,2022-01-07T00:00:00+00:00,2022-01-08T00:00:00+00:00,true,true\n
user2,arn:aws:iam::1234567890:user/user2,2022-01-09T00:00:00+00:00,false,,,2022-01-10T00:00:00+00:00,true,true,2022-01-11T00:00:00+00:00,2022-01-12T00:00:00+00:00,true,2022-01-13T00:00:00+00:00,2022-01-14T00:00:00+00:00,true,true`

var (
	virtualMfaDevices = []types.MFADevice{
		{
			SerialNumber: aws.String("arn:aws:iam::123456789012:mfa/test-user"),
			UserName:     aws.String("test-user-1"),
			EnableDate:   aws.Time(time.Now()),
		},
	}

	mfaDevices = []types.MFADevice{
		{
			SerialNumber: aws.String("MFA-Device"),
			UserName:     aws.String("test-user-2"),
			EnableDate:   aws.Time(time.Now()),
		},
	}

	apiUsers = []types.User{
		{
			UserName:         aws.String("user1"),
			Arn:              aws.String("arn:aws:iam::123456789012:user/user1"),
			CreateDate:       aws.Time(time.Now()),
			PasswordLastUsed: aws.Time(time.Now()),
		},
		{
			UserName:         aws.String("user2"),
			Arn:              aws.String("arn:aws:iam::123456789012:user/user2"),
			CreateDate:       aws.Time(time.Now()),
			PasswordLastUsed: aws.Time(time.Time{}),
		},
	}

	credentialsReportOutput = &iamsdk.GetCredentialReportOutput{
		Content:        []byte(credentialsReportContent),
		GeneratedTime:  &time.Time{},
		ReportFormat:   "text/csv",
		ResultMetadata: middleware.Metadata{},
	}
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
					{credentialsReportOutput, nil}},
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
		t.Run(test.name, func(t *testing.T) {
			p := createProviderFromMockValues(t, test.mocksReturnVals)

			users, err := p.GetUsers(t.Context())

			if !test.wantErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			assert.Len(t, users, test.expectedUsers)
			for i, user := range users {
				assert.Equal(t, test.expected[i].Arn, user.(User).Arn)
				assert.Equal(t, test.expected[i].HasUsed, user.(User).AccessKeys[0].HasUsed)
				assert.Equal(t, test.expected[i].MfaActive, user.(User).MfaActive)
				assert.Equal(t, test.expected[i].hasAttachedPolicy, len(user.(User).AttachedPolicies)+len(user.(User).InlinePolicies) > 0)

				if test.expected[i].MfaActive {
					assert.Equal(t, test.expected[i].IsVirtualMFA, user.(User).MFADevices[0].IsVirtual)
				}
			}
		})
	}
}

func createProviderFromMockValues(t *testing.T, mockReturnValues mocksReturnVals) *Provider {
	mockedClient := MockClient{}
	for funcName, returnValues := range mockReturnValues {
		for _, values := range returnValues {
			mockedClient.On(funcName, mock.Anything, mock.Anything).Return(values...).Once()
		}
	}
	return &Provider{
		log:    testhelper.NewLogger(t),
		client: &mockedClient,
	}
}
