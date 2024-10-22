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
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
)

func TestAzure_Initialize(t *testing.T) {
	baseAzureConfig := config.Config{
		CloudConfig: config.CloudConfig{},
	}
	validAzureConfig := baseAzureConfig

	tests := []struct {
		name                 string
		configProvider       auth.ConfigProviderAPI
		inventoryInitializer azurelib.ProviderInitializerAPI
		cfg                  config.Config
		want                 []string
		wantErr              string
	}{
		{
			name:                 "config provider error",
			cfg:                  baseAzureConfig,
			configProvider:       mockAzureCfgProvider(baseAzureConfig.CloudConfig.Azure, errors.New("some error")),
			inventoryInitializer: mockAzureInventoryInitializerService(nil),
			wantErr:              "some error",
		},
		{
			name:                 "inventory init error",
			cfg:                  validAzureConfig,
			configProvider:       mockAzureCfgProvider(validAzureConfig.CloudConfig.Azure, nil),
			inventoryInitializer: mockAzureInventoryInitializerService(errors.New("some error")),
			wantErr:              "some error",
		},
		{
			name:                 "no error",
			cfg:                  validAzureConfig,
			configProvider:       mockAzureCfgProvider(validAzureConfig.CloudConfig.Azure, nil),
			inventoryInitializer: mockAzureInventoryInitializerService(nil),
			want: []string{
				"azure_cloud_assets_fetcher",
				"azure_cloud_batch_asset_fetcher",
				"azure_cloud_insights_batch_asset_fetcher",
				"azure_cloud_locations_network_watchers_batch_assets_fetcher",
				"azure_security_contacts_assets_fetcher",
			},
		},
		{
			name:                 "no inventory initializer",
			cfg:                  validAzureConfig,
			configProvider:       mockAzureCfgProvider(validAzureConfig.CloudConfig.Azure, nil),
			inventoryInitializer: nil,
			wantErr:              "azure asset inventory is uninitialized",
		},
		{
			name:                 "no config provider",
			cfg:                  validAzureConfig,
			configProvider:       nil,
			inventoryInitializer: mockAzureInventoryInitializerService(nil),
			wantErr:              "azure config provider is uninitialized",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			testInitialize(t, &Azure{
				cfgProvider:         tt.configProvider,
				providerInitializer: tt.inventoryInitializer,
			}, &tt.cfg, tt.wantErr, tt.want)
		})
	}
}

func mockAzureCfgProvider(cfg config.AzureConfig, err error) auth.ConfigProviderAPI {
	cfgProvider := &auth.MockConfigProviderAPI{}
	on := cfgProvider.EXPECT().GetAzureClientConfig(cfg)
	if err == nil {
		on.Return(&auth.AzureFactoryConfig{}, nil)
	} else {
		on.Return(nil, err)
	}
	return cfgProvider
}

func mockAzureInventoryInitializerService(err error) azurelib.ProviderInitializerAPI {
	initializer := &azurelib.MockProviderInitializerAPI{}
	provider := &azurelib.MockProviderAPI{}
	initializerMock := initializer.EXPECT().Init(mock.Anything, mock.Anything)
	if err == nil {
		initializerMock.Return(provider, nil)
	} else {
		initializerMock.Return(nil, err)
	}
	return initializer
}

func TestCalculateFetcherTimeout(t *testing.T) {
	tests := map[string]struct {
		inputPeriod time.Duration
		expected    time.Duration
	}{
		"48h": {
			inputPeriod: 48 * time.Hour,
			expected:    34 * time.Hour,
		},
		"24h": {
			inputPeriod: 24 * time.Hour,
			expected:    17 * time.Hour,
		},
		"3h": {
			inputPeriod: 3 * time.Hour,
			expected:    3 * time.Hour,
		},
		"30m": {
			inputPeriod: 30 * time.Minute,
			expected:    3 * time.Hour,
		},
		"0": {
			inputPeriod: 0,
			expected:    3 * time.Hour,
		},
		"-30m": {
			inputPeriod: -30 * time.Minute,
			expected:    3 * time.Hour,
		},
		"-3h": {
			inputPeriod: -3 * time.Hour,
			expected:    3 * time.Hour,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := calculateFetcherTimeout(tc.inputPeriod)
			require.Equal(t, tc.expected, got)
		})
	}
}
