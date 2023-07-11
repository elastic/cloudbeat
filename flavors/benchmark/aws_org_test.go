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
