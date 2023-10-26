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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

func TestAWSOrg_Initialize(t *testing.T) {
	tests := []struct {
		name             string
		identityProvider awslib.IdentityProviderGetter
		accountProvider  awslib.AccountProviderAPI
		cfg              config.Config
		want             []string
		wantErr          string
	}{
		{
			name:    "nothing initialized",
			wantErr: "aws identity provider is uninitialized",
		},
		{
			name:             "account provider uninitialized",
			identityProvider: mockAwsIdentityProvider(nil),
			accountProvider:  nil,
			wantErr:          "account provider is uninitialized",
		},
		{
			name:             "identity provider error",
			identityProvider: mockAwsIdentityProvider(errors.New("some error")),
			accountProvider:  mockAccountProvider(errors.New("not this error")),
			wantErr:          "some error",
		},
		{
			name:             "account provider error",
			identityProvider: mockAwsIdentityProvider(nil),
			accountProvider:  mockAccountProvider(errors.New("some error")),
			want:             []string{},
		},
		{
			name:             "no error",
			identityProvider: mockAwsIdentityProvider(nil),
			accountProvider:  mockAccountProvider(nil),
			want: []string{
				"123-" + fetching.IAMType,
				"123-" + fetching.KmsType,
				"123-" + fetching.TrailType,
				"123-" + fetching.AwsMonitoringType,
				"123-" + fetching.EC2NetworkingType,
				"123-" + fetching.RdsType,
				"123-" + fetching.S3Type,
				"456-" + fetching.IAMType,
				"456-" + fetching.KmsType,
				"456-" + fetching.TrailType,
				"456-" + fetching.AwsMonitoringType,
				"456-" + fetching.EC2NetworkingType,
				"456-" + fetching.RdsType,
				"456-" + fetching.S3Type,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testInitialize(t, &AWSOrg{
				IdentityProvider: tt.identityProvider,
				AccountProvider:  tt.accountProvider,
			}, &tt.cfg, tt.wantErr, tt.want)
		})
	}
}

func Test_getAwsAccounts(t *testing.T) {
	tests := []struct {
		name            string
		accountProvider awslib.AccountProviderAPI
		rootIdentity    cloud.Identity
		want            []cloud.Identity
		wantErr         string
	}{
		{
			name:            "error",
			accountProvider: mockAccountProvider(errors.New("some error")),
			rootIdentity:    cloud.Identity{Account: "123"},
			wantErr:         "some error",
		},
		{
			name: "success",
			accountProvider: mockAccountProviderWithIdentities([]cloud.Identity{
				{
					Account:      "123",
					AccountAlias: "alias",
				},
				{
					Account:      "456",
					AccountAlias: "alias2",
				},
			}),
			rootIdentity: cloud.Identity{
				Account: "123",
			},
			want: []cloud.Identity{
				{
					Account:      "123",
					AccountAlias: "alias",
				},
				{
					Account:      "456",
					AccountAlias: "alias2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			A := AWSOrg{
				IdentityProvider: nil,
				AccountProvider:  tt.accountProvider,
			}
			got, err := A.getAwsAccounts(context.Background(), nil, aws.Config{}, &tt.rootIdentity)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
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

func mockAccountProvider(err error) *awslib.MockAccountProviderAPI {
	provider := awslib.MockAccountProviderAPI{}
	on := provider.EXPECT().ListAccounts(mock.Anything, mock.Anything, mock.Anything)
	if err == nil {
		on.Return([]cloud.Identity{
			{
				Account:      "123",
				AccountAlias: "some-name",
			},
			{
				Account:      "456",
				AccountAlias: "some-other-name",
			},
		}, nil)
	} else {
		on.Return(nil, err)
	}
	return &provider
}

func mockAccountProviderWithIdentities(identities []cloud.Identity) *awslib.MockAccountProviderAPI {
	provider := awslib.MockAccountProviderAPI{}
	provider.EXPECT().ListAccounts(mock.Anything, mock.Anything, mock.Anything).Return(identities, nil)
	return &provider
}
