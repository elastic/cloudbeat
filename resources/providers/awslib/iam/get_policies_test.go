package iam

import (
	"context"
	"errors"
	iamsdk "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvider_GetPolicies(t *testing.T) {
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
			name: "Error listing policies",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {{nil, errors.New("some error")}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Policies with missing fields",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {
					{
						&iamsdk.ListPoliciesOutput{
							Policies: []types.Policy{{Arn: nil},
								{
									Arn:              aws.String("some-arn"),
									DefaultVersionId: nil,
								},
							},
						},
						nil,
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Error getting policy version",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {
					{
						&iamsdk.ListPoliciesOutput{
							Policies: []types.Policy{{Arn: nil},
								{
									Arn:              aws.String("some-arn"),
									DefaultVersionId: aws.String("some-version"),
								},
							},
						},
						nil,
					},
				},
				"GetPolicyVersion": {{nil, errors.New("some error")}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Return 1 policy",
			mockReturnValues: mocksReturnVals{
				"ListPolicies": {
					{
						&iamsdk.ListPoliciesOutput{
							Policies: []types.Policy{{Arn: nil},
								{
									Arn:              aws.String("some-arn"),
									DefaultVersionId: aws.String("some-version"),
								},
							},
						},
						nil,
					},
				},
				"GetPolicyVersion": {{&iamsdk.GetPolicyVersionOutput{
					PolicyVersion: &types.PolicyVersion{
						Document: aws.String("%7B%22hello%22%3A%20%22world%22%7D"),
					},
				}, nil}},
			},
			want: []awslib.AwsResource{
				Policy{
					Policy: types.Policy{
						Arn:              aws.String("some-arn"),
						DefaultVersionId: aws.String("some-version"),
					},
					Document: map[string]interface{}{"hello": "world"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createProviderFromMockValues(tt.mockReturnValues)

			got, err := p.GetPolicies(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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
		want          map[string]interface{}
		wantErr       string
	}{
		{
			name:          "Check for nil policy version",
			policyVersion: nil,
			want:          nil,
		},
		{
			name: "Check for nil document",
			policyVersion: &types.PolicyVersion{
				Document: nil,
			},
			want: nil,
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
			policyVersion: docToPolicy("%7B%22hello%22%3A%20%22world%22%7D"), // {"hello": "world"}
			want:          map[string]interface{}{"hello": "world"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodePolicyDocument(tt.policyVersion)
			if tt.wantErr != "" {
				assert.ErrorContainsf(t, err, tt.wantErr, "decodePolicyDocument(%v)", tt.policyVersion)
			} else {
				assert.NoError(t, err, "decodePolicyDocument(%v)", tt.policyVersion)
			}
			assert.Equalf(t, tt.want, got, "decodePolicyDocument(%v)", tt.policyVersion)
		})
	}
}
