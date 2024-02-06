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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/auth"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/governance"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type ProviderAPI interface {
	inventory.ResourceGraphProviderAPI
	inventory.SQLProviderAPI
	inventory.MysqlProviderAPI
	inventory.StorageAccountProviderAPI
	inventory.PostgresqlProviderAPI
	inventory.KeyVaultProviderAPI
	inventory.SubscriptionProviderAPI
	inventory.SecurityContactsProviderAPI
	governance.ProviderAPI
}

type ProviderInitializer struct{}

type ProviderInitializerAPI interface {
	// Init initializes the Azure clients
	Init(log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ProviderAPI, error)
}

func (p *ProviderInitializer) Init(log *logp.Logger, azureConfig auth.AzureFactoryConfig) (ProviderAPI, error) {
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

	genericARMClient, err := arm.NewClient("armsecurity-custom", "v0.0.1", azureConfig.Credentials, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init generic arm client: %w", err)
	}

	resourceGraphProvider := inventory.NewResourceGraphProvider(log, resourceGraphClientFactory)
	return &provider{
		ResourceGraphProviderAPI:    resourceGraphProvider,
		SQLProviderAPI:              inventory.NewSQLProvider(log, azureConfig.Credentials),
		MysqlProviderAPI:            inventory.NewMysqlProvider(log, azureConfig.Credentials),
		PostgresqlProviderAPI:       inventory.NewPostgresqlProvider(log, azureConfig.Credentials),
		StorageAccountProviderAPI:   inventory.NewStorageAccountProvider(log, diagnosticSettingsClient, azureConfig.Credentials),
		KeyVaultProviderAPI:         inventory.NewKeyVaultProvider(log, azureConfig.Credentials),
		SubscriptionProviderAPI:     inventory.NewSubscriptionProvider(log, azureConfig.Credentials),
		SecurityContactsProviderAPI: inventory.NewSecurityContacts(log, azureConfig.Credentials, genericARMClient),
		ProviderAPI:                 governance.NewProvider(log, resourceGraphProvider),
	}, nil
}

type provider struct {
	inventory.ResourceGraphProviderAPI
	inventory.SQLProviderAPI
	inventory.MysqlProviderAPI
	inventory.StorageAccountProviderAPI
	inventory.PostgresqlProviderAPI
	inventory.KeyVaultProviderAPI
	inventory.SubscriptionProviderAPI
	inventory.SecurityContactsProviderAPI
	governance.ProviderAPI
}
