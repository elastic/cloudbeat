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
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/identity"
)

func TestGCP_Initialize(t *testing.T) {
	baseGcpConfig := config.Config{
		CloudConfig: config.CloudConfig{
			Gcp: config.GcpConfig{
				ProjectId:    "some-project",
				GcpClientOpt: config.GcpClientOpt{},
			},
		},
	}
	validGcpConfig := baseGcpConfig
	validGcpConfig.CloudConfig.Gcp.CredentialsJSON = `{
               "type": "authorized_user"
       }`

	tests := []struct {
		name             string
		identityProvider identity.ProviderGetter
		configProvider   auth.ConfigProviderAPI
		cfg              config.Config
		want             []string
		wantErr          string
	}{
		{
			name:    "nothing initialized",
			cfg:     baseGcpConfig,
			wantErr: "gcp identity provider is uninitialized",
		},
		{
			name:             "missing credentials options, fallback to using ADC",
			cfg:              baseGcpConfig,
			identityProvider: mockGcpIdentityProvider(nil),
			configProvider:   mockGcpCfgProvider(nil),
			want: []string{
				"gcp_cloud_assets_fetcher",
			},
		},
		{
			name:             "config provider error",
			cfg:              baseGcpConfig,
			identityProvider: mockGcpIdentityProvider(nil),
			configProvider:   mockGcpCfgProvider(errors.New("some error")),
			wantErr:          "some error",
		},
		{
			name:             "identity provider error",
			cfg:              validGcpConfig,
			identityProvider: mockGcpIdentityProvider(errors.New("some error")),
			configProvider:   mockGcpCfgProvider(nil),
			wantErr:          "some error",
		},
		{
			name:             "no error",
			cfg:              validGcpConfig,
			identityProvider: mockGcpIdentityProvider(nil),
			configProvider:   mockGcpCfgProvider(nil),
			want: []string{
				"gcp_cloud_assets_fetcher",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testInitialize(t, &GCP{
				IdentityProvider: tt.identityProvider,
				CfgProvider:      tt.configProvider,
			}, &tt.cfg, tt.wantErr, tt.want)
		})
	}
}

func mockGcpIdentityProvider(err error) *identity.MockProviderGetter {
	identityProvider := &identity.MockProviderGetter{}
	on := identityProvider.EXPECT().GetIdentity(mock.Anything, mock.Anything, mock.Anything)
	if err == nil {
		on.Return(
			&cloud.Identity{
				Provider:    "gcp",
				ProjectId:   "test-project-id",
				ProjectName: "test-project-name",
			},
			nil,
		)
	} else {
		on.Return(nil, err)
	}
	return identityProvider
}

func mockGcpCfgProvider(err error) auth.ConfigProviderAPI {
	cfgProvider := &auth.MockConfigProviderAPI{}
	on := cfgProvider.EXPECT().GetGcpClientConfig(mock.Anything, mock.Anything)
	if err == nil {
		on.Return(
			[]option.ClientOption{},
			nil,
		)
	} else {
		on.Return(nil, err)
	}
	return cfgProvider
}
