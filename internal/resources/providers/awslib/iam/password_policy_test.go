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
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

func Test_GetPasswordPolicy(t *testing.T) {
	tcs := []struct {
		tcName    string
		expectErr bool
		mockErr   error
		mockRes   *iam.GetAccountPasswordPolicyOutput
		expected  awslib.AwsResource
	}{
		{
			tcName:    "Happy path",
			expectErr: false,
			mockErr:   nil,
			mockRes: &iam.GetAccountPasswordPolicyOutput{
				PasswordPolicy: &types.PasswordPolicy{
					PasswordReusePrevention:    ptr.Int32(24),
					MaxPasswordAge:             ptr.Int32(90),
					MinimumPasswordLength:      ptr.Int32(16),
					RequireLowercaseCharacters: true,
					RequireUppercaseCharacters: true,
					RequireNumbers:             true,
					RequireSymbols:             true,
				},
			},
			expected: &PasswordPolicy{
				ReusePreventionCount: 24,
				RequireLowercase:     true,
				RequireUppercase:     true,
				RequireNumbers:       true,
				RequireSymbols:       true,
				MaxAgeDays:           90,
				MinimumLength:        16,
			},
		},

		{
			tcName:    "NoSuchEntityException",
			expectErr: false,
			mockErr:   &types.NoSuchEntityException{},
			mockRes:   nil,
			expected:  &PasswordPolicy{},
		},

		{
			tcName:    "MalformedPolicyDocumentException",
			expectErr: true,
			mockErr:   &types.MalformedPolicyDocumentException{},
			mockRes:   nil,
			expected:  nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.tcName, func(t *testing.T) {
			ctx := context.Background()
			log := clog.NewLogger("test")
			client := &MockClient{}
			provider := Provider{
				log:                   log,
				client:                client,
				accessAnalyzerClients: nil,
			}

			client.EXPECT().GetAccountPasswordPolicy(ctx, &iam.GetAccountPasswordPolicyInput{}).Return(tc.mockRes, tc.mockErr)
			pp, err := provider.GetPasswordPolicy(ctx)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.expected, pp)
		})
	}
}
