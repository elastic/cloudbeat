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

package azurelib

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type ProviderAPI interface {
	governance.ProviderAPI
	inventory.AppServiceProviderAPI
	inventory.KeyVaultProviderAPI
	inventory.MysqlProviderAPI
	inventory.PostgresqlProviderAPI
	inventory.ResourceGraphProviderAPI
	inventory.SQLProviderAPI
	inventory.SecurityContactsProviderAPI
	inventory.StorageAccountProviderAPI
	inventory.SubscriptionProviderAPI
}

type ProviderInitializer struct{}

type ProviderInitializerAPI interface {
	// Init initializes the Azure clients
	Init(log *clog.Logger, azureConfig auth.AzureFactoryConfig) (ProviderAPI, error)
}

func (p *ProviderInitializer) Init(log *clog.Logger, azureConfig auth.AzureFactoryConfig) (ProviderAPI, error) {
	log = log.Named("azure")

	factory, err := armresourcegraph.NewClientFactory(azureConfig.Credentials, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize resource graph factory: %w", err)
	}
	resourceGraphClientFactory := factory.NewClient()

	diagnosticSettingsClient, err := armmonitor.NewDiagnosticSettingsClient(azureConfig.Credentials, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init monitor client: %w", err)
	}

	resourceGraphProvider := inventory.NewResourceGraphProvider(log, resourceGraphClientFactory)
	return &provider{
		AppServiceProviderAPI:       inventory.NewAppServiceProvider(log, azureConfig.Credentials),
		KeyVaultProviderAPI:         inventory.NewKeyVaultProvider(log, diagnosticSettingsClient, azureConfig.Credentials),
		MysqlProviderAPI:            inventory.NewMysqlProvider(log, azureConfig.Credentials),
		PostgresqlProviderAPI:       inventory.NewPostgresqlProvider(log, azureConfig.Credentials),
		ProviderAPI:                 governance.NewProvider(log, resourceGraphProvider),
		ResourceGraphProviderAPI:    resourceGraphProvider,
		SQLProviderAPI:              inventory.NewSQLProvider(log, azureConfig.Credentials),
		SecurityContactsProviderAPI: inventory.NewSecurityContacts(log, azureConfig.Credentials),
		StorageAccountProviderAPI:   inventory.NewStorageAccountProvider(log, diagnosticSettingsClient, azureConfig.Credentials),
		SubscriptionProviderAPI:     inventory.NewSubscriptionProvider(log, azureConfig.Credentials),
	}, nil
}

type provider struct {
	governance.ProviderAPI
	inventory.AppServiceProviderAPI
	inventory.KeyVaultProviderAPI
	inventory.MysqlProviderAPI
	inventory.PostgresqlProviderAPI
	inventory.ResourceGraphProviderAPI
	inventory.SQLProviderAPI
	inventory.SecurityContactsProviderAPI
	inventory.StorageAccountProviderAPI
	inventory.SubscriptionProviderAPI
}
