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

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/identity"
	"github.com/elastic/cloudbeat/resources/providers/gcplib/inventory"
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
		name                 string
		identityProvider     identity.ProviderGetter
		configProvider       auth.ConfigProviderAPI
		inventoryInitializer inventory.ProviderInitializerAPI
		cfg                  config.Config
		want                 []string
		wantErr              string
	}{
		{
			name:    "nothing initialized",
			cfg:     baseGcpConfig,
			wantErr: "gcp identity provider is uninitialized",
		},
		{
			name:                 "missing credentials options, fallback to using ADC",
			cfg:                  baseGcpConfig,
			identityProvider:     mockGcpIdentityProvider(nil),
			configProvider:       mockGcpCfgProvider(nil),
			inventoryInitializer: mockInventoryInitializerService(nil),
			want: []string{
				"gcp_cloud_assets_fetcher",
				"gcp_monitoring_fetcher",
			},
		},
		{
			name:                 "config provider error",
			cfg:                  baseGcpConfig,
			identityProvider:     mockGcpIdentityProvider(nil),
			configProvider:       mockGcpCfgProvider(errors.New("some error")),
			inventoryInitializer: mockInventoryInitializerService(nil),
			wantErr:              "some error",
		},
		{
			name:                 "identity provider error",
			cfg:                  validGcpConfig,
			identityProvider:     mockGcpIdentityProvider(errors.New("some error")),
			configProvider:       mockGcpCfgProvider(nil),
			inventoryInitializer: mockInventoryInitializerService(nil),
			wantErr:              "some error",
		},
		{
			name:                 "inventory init error",
			cfg:                  validGcpConfig,
			identityProvider:     mockGcpIdentityProvider(nil),
			configProvider:       mockGcpCfgProvider(nil),
			inventoryInitializer: mockInventoryInitializerService(errors.New("some error")),
			wantErr:              "some error",
		},
		{
			name:                 "no error",
			cfg:                  validGcpConfig,
			identityProvider:     mockGcpIdentityProvider(nil),
			configProvider:       mockGcpCfgProvider(nil),
			inventoryInitializer: mockInventoryInitializerService(nil),
			want: []string{
				"gcp_cloud_assets_fetcher",
				"gcp_monitoring_fetcher",
			},
		},
		{
			name:                 "no identity provider",
			cfg:                  validGcpConfig,
			identityProvider:     nil,
			configProvider:       mockGcpCfgProvider(nil),
			inventoryInitializer: mockInventoryInitializerService(nil),
			wantErr:              "gcp identity provider is uninitialized",
		},
		{
			name:                 "no inventory initializer",
			cfg:                  validGcpConfig,
			identityProvider:     mockGcpIdentityProvider(nil),
			configProvider:       mockGcpCfgProvider(nil),
			inventoryInitializer: nil,
			wantErr:              "gcp asset inventory is uninitialized",
		},
		{
			name:                 "no config provider",
			cfg:                  validGcpConfig,
			identityProvider:     mockGcpIdentityProvider(nil),
			configProvider:       nil,
			inventoryInitializer: mockInventoryInitializerService(nil),
			wantErr:              "gcp config provider is uninitialized",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			testInitialize(t, &GCP{
				IdentityProvider:     tt.identityProvider,
				CfgProvider:          tt.configProvider,
				inventoryInitializer: tt.inventoryInitializer,
			}, &tt.cfg, tt.wantErr, tt.want)
		})
	}
}

func mockGcpIdentityProvider(err error) *identity.MockProviderGetter {
	identityProvider := &identity.MockProviderGetter{}
	on := identityProvider.EXPECT().GetIdentity(mock.Anything, mock.Anything)
	if err == nil {
		on.Return(
			&cloud.Identity{
				Provider:     "gcp",
				Account:      "test-project-id",
				AccountAlias: "test-project-name",
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
	on := cfgProvider.EXPECT().GetGcpClientConfig(mock.Anything, mock.Anything, mock.Anything)
	if err == nil {
		on.Return(
			&auth.GcpFactoryConfig{},
			nil,
		)
	} else {
		on.Return(nil, err)
	}
	return cfgProvider
}

func mockInventoryInitializerService(err error) inventory.ProviderInitializerAPI {
	initializer := &inventory.MockProviderInitializerAPI{}
	inventoryService := &inventory.MockServiceAPI{}
	initializerMock := initializer.EXPECT().Init(mock.Anything, mock.Anything, mock.Anything)
	if err == nil {
		initializerMock.Return(
			inventoryService,
			nil,
		)
	} else {
		initializerMock.Return(nil, err)
	}
	return initializer
}
