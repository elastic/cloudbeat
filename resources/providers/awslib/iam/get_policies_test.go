package iam

import (
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
