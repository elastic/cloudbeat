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
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mocksReturnVals map[string][][]any

type expectedResult struct {
	HasUsed      bool
	IsVirtualMFA bool
	Arn          string
	MfaActive    interface{}
}

func Test_GetUsers(t *testing.T) {
	tests := []struct {
		name            string
		mocksReturnVals mocksReturnVals
		expected        []expectedResult
		wantErr         bool
	}{
		{
			name: "Test 1: Should not return users",
			mocksReturnVals: mocksReturnVals{
				"ListUsers":                {{&iamsdk.ListUsersOutput{}, nil}},
				"ListMFADevices":           {nil, nil},
				"GenerateCredentialReport": {nil, nil},
				"GetCredentialReport":      {{CredentialsReportOutput, nil}},
			},
			expected: []expectedResult{
				{
					HasUsed:      false,
					IsVirtualMFA: false,
					Arn:          "",
					MfaActive:    true,
				},
			},
			wantErr: false,
		},
		{
			name: "Test 2: Should return two users",
			mocksReturnVals: mocksReturnVals{
				"ListUsers": {{&iamsdk.ListUsersOutput{Users: apiUsers}, nil}},
				"ListMFADevices": {
					{&iamsdk.ListMFADevicesOutput{MFADevices: virtualMfaDevices}, nil},
					{&iamsdk.ListMFADevicesOutput{MFADevices: mfaDevices}, nil},
					{nil, errors.New("no such user - root account")},
				},
				"GenerateCredentialReport": {nil, nil},
				"GetCredentialReport":      {{CredentialsReportOutput, nil}},
			},
			expected: []expectedResult{
				{
					HasUsed:      true,
					MfaActive:    true,
					IsVirtualMFA: true,
					Arn:          "arn:aws:iam::123456789012:user/user1",
				},
				{
					HasUsed:      true,
					MfaActive:    true,
					IsVirtualMFA: false,
					Arn:          "arn:aws:iam::123456789012:user/user2",
				},
				{
					HasUsed:      true,
					MfaActive:    false,
					IsVirtualMFA: false,
					Arn:          "arn:aws:iam::1234567890:root",
				},
			},
			wantErr: false,
		},
		{
			name: "Test 3: Error should not return users",
			mocksReturnVals: mocksReturnVals{
				"ListUsers": {{&iamsdk.ListUsersOutput{}, errors.New("fail to list iam users")}},
			},
			expected: []expectedResult{
				{
					HasUsed:      false,
					MfaActive:    true,
					IsVirtualMFA: false,
					Arn:          "",
				},
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		mockedClient := &MockClient{}
		for funcName, returnVals := range test.mocksReturnVals {
			for _, vals := range returnVals {
				mockedClient.On(funcName, mock.Anything, mock.Anything).Return(vals...).Once()
			}
		}

		p := Provider{
			client: mockedClient,
			log:    logp.NewLogger("iam-provider"),
		}

		users, err := p.GetUsers(context.TODO())

		if !test.wantErr {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}

		for i, user := range users {
			assert.Equal(t, test.expected[i].Arn, user.(User).Arn)
			assert.Equal(t, test.expected[i].HasUsed, user.(User).AccessKeys[0].HasUsed)
			assert.Equal(t, test.expected[i].MfaActive, user.(User).MfaActive)
			assert.Equal(t, test.expected[i].IsVirtualMFA, user.(User).MFADevices[0].IsVirtual)
		}
	}
}
