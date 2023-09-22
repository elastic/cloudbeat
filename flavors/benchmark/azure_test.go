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
	"github.com/elastic/cloudbeat/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

func TestAzure_Initialize(t *testing.T) {
	baseAzureConfig := config.Config{
		CloudConfig: config.CloudConfig{},
	}
	validAzureConfig := baseAzureConfig

	tests := []struct {
		name                 string
		configProvider       auth.ConfigProviderAPI
		inventoryInitializer inventory.ProviderInitializerAPI
		cfg                  config.Config
		want                 []string
		wantErr              string
	}{
		{
			name:                 "config provider error",
			cfg:                  baseAzureConfig,
			configProvider:       mockAzureCfgProvider(errors.New("some error")),
			inventoryInitializer: mockAzureInventoryInitializerService(nil),
			wantErr:              "some error",
		},
		{
			name:                 "inventory init error",
			cfg:                  validAzureConfig,
			configProvider:       mockAzureCfgProvider(nil),
			inventoryInitializer: mockAzureInventoryInitializerService(errors.New("some error")),
			wantErr:              "some error",
		},
		{
			name:                 "no error",
			cfg:                  validAzureConfig,
			configProvider:       mockAzureCfgProvider(nil),
			inventoryInitializer: mockAzureInventoryInitializerService(nil),
			want: []string{
				"azure_cloud_assets_fetcher",
			},
		},
		{
			name:                 "no inventory initializer",
			cfg:                  validAzureConfig,
			configProvider:       mockAzureCfgProvider(nil),
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
				CfgProvider:          tt.configProvider,
				inventoryInitializer: tt.inventoryInitializer,
			}, &tt.cfg, tt.wantErr, tt.want)
		})
	}
}

func mockAzureCfgProvider(err error) auth.ConfigProviderAPI {
	cfgProvider := &auth.MockConfigProviderAPI{}
	on := cfgProvider.EXPECT().GetAzureClientConfig()
	if err == nil {
		on.Return(
			&auth.AzureFactoryConfig{},
			nil,
		)
	} else {
		on.Return(nil, err)
	}
	return cfgProvider
}

func mockAzureInventoryInitializerService(err error) inventory.ProviderInitializerAPI {
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
