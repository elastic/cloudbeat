package iam

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/ptr"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/require"
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
			log := logp.NewLogger("test")
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
