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
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestProvider_GetPolicies(t *testing.T) {
	supportPolicy := validSupportAccessPolicy()
	supportPolicyOut := &iamsdk.GetPolicyOutput{Policy: &supportPolicy}

	tests := []struct {
		name             string
		mockReturnValues mocksReturnVals
		want             []awslib.AwsResource
		wantErr          bool
	}{
		{
			name: "Error listing policies",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {{nil, errors.New("some error")}},
			},
			wantErr: true,
		},
		{
			name: "Error in GetPolicy",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {
					{
						&iamsdk.ListPoliciesOutput{Policies: []types.Policy{validPolicy()}},
						nil,
					},
				},
				"GetPolicyVersion": {{validGetPolicyVersionOutput(), nil}},
				"GetPolicy": {
					{
						nil,
						errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Return just the support policy policy",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {
					{
						&iamsdk.ListPoliciesOutput{},
						nil,
					},
				},
				"GetPolicyVersion": {{validGetPolicyVersionOutput(), nil}},
				"GetPolicy": {
					{
						supportPolicyOut,
						nil,
					},
				},
				"ListEntitiesForPolicy": {
					{
						&iamsdk.ListEntitiesForPolicyOutput{
							IsTruncated: false,
							PolicyRoles: validPolicyRoles(),
						},
						nil,
					},
				},
			},
			want: []awslib.AwsResource{
				Policy{
					Policy:   validSupportAccessPolicy(),
					Document: validGetPolicyVersionOutputDecoded(),
					Roles:    validPolicyRoles(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createProviderFromMockValues(t, tt.mockReturnValues)

			got, err := p.GetPolicies(t.Context())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProvider_getPolicies(t *testing.T) {
	tests := []struct {
		name             string
		mockReturnValues mocksReturnVals
		want             []awslib.AwsResource
		wantErr          bool
	}{
		{
			name: "No policies",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {{&iamsdk.ListPoliciesOutput{}, nil}},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Ignore AWSSupportPolicy",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {
					{
						&iamsdk.ListPoliciesOutput{
							Policies: []types.Policy{validSupportAccessPolicy()},
						},
						nil,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Error listing policies",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {{nil, errors.New("some error")}},
			},
			wantErr: true,
		},
		{
			name: "Policies with missing fields",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {
					{
						&iamsdk.ListPoliciesOutput{
							Policies: []types.Policy{
								{},
								invalidPolicyWithoutVersion(),
							},
						},
						nil,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Error getting policy version",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {
					{
						&iamsdk.ListPoliciesOutput{
							Policies: []types.Policy{validPolicy()},
						},
						nil,
					},
				},
				"GetPolicyVersion": {{nil, errors.New("some error")}},
			},
			wantErr: true,
		},
		{
			name: "Return 1 policy",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {
					{
						&iamsdk.ListPoliciesOutput{
							Policies: []types.Policy{validPolicy()},
						},
						nil,
					},
				},
				"GetPolicyVersion": {{validGetPolicyVersionOutput(), nil}},
			},
			want: []awslib.AwsResource{
				Policy{
					Policy:   validPolicy(),
					Document: validGetPolicyVersionOutputDecoded(),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createProviderFromMockValues(t, tt.mockReturnValues)

			got, err := p.getPolicies(t.Context())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestProvider_getSupportPolicy(t *testing.T) {
	policy := validSupportAccessPolicy()
	policyOut := &iamsdk.GetPolicyOutput{Policy: &policy}

	tests := []struct {
		name             string
		mockReturnValues mocksReturnVals
		want             awslib.AwsResource
		wantErr          bool
	}{
		{
			name: "Error in GetPolicy",
			mockReturnValues: mocksReturnVals{
				"GetPolicy": {
					{
						nil,
						errors.New("some error"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Error in ListEntitiesForPolicy",
			mockReturnValues: mocksReturnVals{
				"GetPolicy": {
					{
						policyOut,
						nil,
					},
				},
				"ListEntitiesForPolicy": {
					{
						nil,
						errors.New("some error"),
					},
				},
				"GetPolicyVersion": {{validGetPolicyVersionOutput(), nil}},
			},
			wantErr: true,
		},
		{
			name: "Empty",
			mockReturnValues: mocksReturnVals{
				"GetPolicy": {
					{
						policyOut,
						nil,
					},
				},
				"ListEntitiesForPolicy": {
					{
						&iamsdk.ListEntitiesForPolicyOutput{
							IsTruncated: false,
							PolicyRoles: []types.PolicyRole{},
						},
						nil,
					},
				},
				"GetPolicyVersion": {{validGetPolicyVersionOutput(), nil}},
			},
			want: Policy{
				Policy:   validSupportAccessPolicy(),
				Document: validGetPolicyVersionOutputDecoded(),
				Roles:    []types.PolicyRole{},
			},
			wantErr: false,
		},
		{
			name: "Success",
			mockReturnValues: mocksReturnVals{
				"GetPolicy": {
					{
						policyOut,
						nil,
					},
				},
				"ListEntitiesForPolicy": {
					{
						&iamsdk.ListEntitiesForPolicyOutput{
							IsTruncated: false,
							PolicyRoles: validPolicyRoles(),
						},
						nil,
					},
				},
				"GetPolicyVersion": {{validGetPolicyVersionOutput(), nil}},
			},
			want: Policy{
				Policy:   validSupportAccessPolicy(),
				Document: validGetPolicyVersionOutputDecoded(),
				Roles:    validPolicyRoles(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createProviderFromMockValues(t, tt.mockReturnValues)

			got, err := p.getSupportPolicy(t.Context())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_decodePolicyDocument(t *testing.T) {
	docToPolicy := func(document string) *types.PolicyVersion {
		return &types.PolicyVersion{
			Document: aws.String(document),
		}
	}

	tests := []struct {
		name          string
		policyVersion *types.PolicyVersion
		want          map[string]any
		wantErr       string
	}{
		{
			name:          "Check for nil policy version",
			policyVersion: nil,
			wantErr:       "invalid policy version",
		},
		{
			name: "Check for nil document",
			policyVersion: &types.PolicyVersion{
				Document: nil,
			},
			wantErr: "invalid policy version",
		},
		{
			name:          "Invalid JSON",
			policyVersion: docToPolicy("xxx"),
			want:          nil,
			wantErr:       "failed to unmarshal",
		},
		{
			name:          "Invalid RFC 3986",
			policyVersion: docToPolicy("hello%world"),
			want:          nil,
			wantErr:       "failed to unescape",
		},
		{
			name:          "Success",
			policyVersion: validGetPolicyVersionOutput().PolicyVersion,
			want:          validGetPolicyVersionOutputDecoded(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodePolicyDocument(tt.policyVersion)
			if tt.wantErr != "" {
				require.ErrorContainsf(t, err, tt.wantErr, "decodePolicyDocument(%v)", tt.policyVersion)
			} else {
				require.NoError(t, err, "decodePolicyDocument(%v)", tt.policyVersion)
			}
			assert.Equalf(t, tt.want, got, "decodePolicyDocument(%v)", tt.policyVersion)
		})
	}
}

func validSupportAccessPolicy() types.Policy {
	return types.Policy{Arn: aws.String("arn:aws:iam::aws:policy/AWSSupportAccess"), DefaultVersionId: aws.String("some-version")}
}

func validPolicy() types.Policy {
	return types.Policy{Arn: aws.String("some-arn"), DefaultVersionId: aws.String("some-version")}
}

func invalidPolicyWithoutVersion() types.Policy {
	return types.Policy{Arn: aws.String("some-arn")}
}

func validGetPolicyVersionOutput() *iamsdk.GetPolicyVersionOutput {
	return &iamsdk.GetPolicyVersionOutput{
		PolicyVersion: &types.PolicyVersion{
			Document: aws.String("%7B%22hello%22%3A%20%22world%22%7D"),
		},
	}
}

func validGetPolicyVersionOutputDecoded() map[string]any {
	return map[string]any{"hello": "world"}
}

func validPolicyRoles() []types.PolicyRole {
	return []types.PolicyRole{
		{
			RoleId:   aws.String("role-id"),
			RoleName: aws.String("role-name"),
		},
		{
			RoleId:   aws.String("role 2"),
			RoleName: aws.String("name 2"),
		},
	}
}

func TestProvider_supportAccessPolicyARN(t *testing.T) {
	tests := []struct {
		partition string
		want      string
	}{
		{partition: "", want: "arn:aws:iam::aws:policy/AWSSupportAccess"},
		{partition: awslib.DefaultPartition, want: "arn:aws:iam::aws:policy/AWSSupportAccess"},
		{partition: awslib.GovCloudPartition, want: "arn:aws-us-gov:iam::aws:policy/AWSSupportAccess"},
		{partition: awslib.ChinaPartition, want: "arn:aws-cn:iam::aws:policy/AWSSupportAccess"},
	}
	for _, tt := range tests {
		t.Run(tt.partition, func(t *testing.T) {
			p := Provider{partition: tt.partition}
			assert.Equal(t, tt.want, p.supportAccessPolicyARN())
		})
	}
}

func TestProvider_GetPolicies_GovCloud(t *testing.T) {
	const govSupportPolicyARN = "arn:aws-us-gov:iam::aws:policy/AWSSupportAccess"
	govSupportPolicy := types.Policy{Arn: aws.String(govSupportPolicyARN), DefaultVersionId: aws.String("some-version")}

	mockedClient := &MockClient{}
	// ListPolicies also returns the support policy. It must not be in the result two times.
	mockedClient.On("ListPolicies", mock.Anything, mock.Anything).Return(
		&iamsdk.ListPoliciesOutput{Policies: []types.Policy{validPolicy(), govSupportPolicy}}, nil,
	).Once()
	mockedClient.On("GetPolicyVersion", mock.Anything, mock.Anything).Return(validGetPolicyVersionOutput(), nil).Twice()
	// The GetPolicy and ListEntitiesForPolicy calls must use the GovCloud ARN.
	mockedClient.On("GetPolicy", mock.Anything, &iamsdk.GetPolicyInput{PolicyArn: aws.String(govSupportPolicyARN)}).Return(
		&iamsdk.GetPolicyOutput{Policy: &govSupportPolicy}, nil,
	).Once()
	mockedClient.On("ListEntitiesForPolicy", mock.Anything, &iamsdk.ListEntitiesForPolicyInput{
		PolicyArn:    aws.String(govSupportPolicyARN),
		EntityFilter: types.EntityTypeRole,
	}).Return(&iamsdk.ListEntitiesForPolicyOutput{PolicyRoles: validPolicyRoles()}, nil).Once()

	p := &Provider{
		log:       testhelper.NewLogger(t),
		client:    mockedClient,
		partition: awslib.GovCloudPartition,
	}

	got, err := p.GetPolicies(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []awslib.AwsResource{
		Policy{Policy: validPolicy(), Document: validGetPolicyVersionOutputDecoded()},
		Policy{Policy: govSupportPolicy, Document: validGetPolicyVersionOutputDecoded(), Roles: validPolicyRoles()},
	}, got)
	mockedClient.AssertExpectations(t)
}
