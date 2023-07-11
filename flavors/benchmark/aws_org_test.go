package benchmark

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

func Test_getAwsAccounts(t *testing.T) {
	tests := []struct {
		name            string
		accountProvider awslib.AccountProviderAPI
		rootIdentity    awslib.Identity
		want            []awslib.Identity
		wantErr         string
	}{
		{
			name:            "error",
			accountProvider: mockAccountProvider(errors.New("some error")),
			rootIdentity:    awslib.Identity{Account: "123"},
			wantErr:         "some error",
		},
		{
			name: "",
			accountProvider: mockAccountProviderWithIdentities([]awslib.Identity{
				{
					Account: "123",
				},
				{
					Account: "456",
					Alias:   "alias2",
				},
			}),
			rootIdentity: awslib.Identity{
				Account: "123",
				Alias:   "alias",
			},
			want: []awslib.Identity{
				{
					Account: "123",
					Alias:   "alias",
				},
				{
					Account: "456",
					Alias:   "alias2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAwsAccounts(
				context.Background(),
				aws.Config{},
				&Dependencies{
					AwsCfgProvider:           nil,
					AwsIdentityProvider:      nil,
					AwsAccountProvider:       tt.accountProvider,
					KubernetesClientProvider: nil,
					AwsMetadataProvider:      nil,
					EksClusterNameProvider:   nil,
				},
				&tt.rootIdentity,
			)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, len(tt.want))

			for i, account := range got {
				assert.Equal(t, tt.want[i], account.Identity)
				assert.IsType(t, &aws.CredentialsCache{}, account.Credentials)
			}
		})
	}
}
