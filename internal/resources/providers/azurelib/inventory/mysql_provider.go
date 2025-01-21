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

package inventory

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysqlflexibleservers"

	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

const (
	mysqlConfigurationTLSVersion = "tls_version"
)

type mysqlAzureClientWrapper struct {
	AssetFlexibleConfiguration func(ctx context.Context, subID, resourceGroup, serverName, configName string, clientOptions *arm.ClientOptions, options *armmysqlflexibleservers.ConfigurationsClientGetOptions) (armmysqlflexibleservers.ConfigurationsClientGetResponse, error)
}

type MysqlProviderAPI interface {
	// GetFlexibleTLSVersionConfiguration fetches SSL Configuration for flexible mysql servers
	// We are fetching specifically SSL configuration only. There is, though a bulk configurations call
	// in the SDK. If more configurations for flexible mysql servers are needed, it's good switch to the bulk call
	// instead of making singular calls
	GetFlexibleTLSVersionConfiguration(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error)
}

type mysqlProvider struct {
	client mysqlAzureClientWrapper
	log    *clog.Logger //nolint:unused
}

func NewMysqlProvider(log *clog.Logger, credentials azcore.TokenCredential) MysqlProviderAPI {
	// We wrap the client, so we can mock it in tests
	client := mysqlAzureClientWrapper{
		AssetFlexibleConfiguration: func(ctx context.Context, subID, resourceGroup, serverName, configName string, clientOptions *arm.ClientOptions, options *armmysqlflexibleservers.ConfigurationsClientGetOptions) (armmysqlflexibleservers.ConfigurationsClientGetResponse, error) {
			cl, err := armmysqlflexibleservers.NewConfigurationsClient(subID, credentials, clientOptions)
			if err != nil {
				return armmysqlflexibleservers.ConfigurationsClientGetResponse{}, err
			}

			return cl.Get(ctx, resourceGroup, serverName, configName, options)
		},
	}

	return &mysqlProvider{
		client: client,
		log:    log,
	}
}

func (p *mysqlProvider) GetFlexibleTLSVersionConfiguration(ctx context.Context, subID, resourceGroup, serverName string) ([]AzureAsset, error) {
	tlsVersion, err := p.client.AssetFlexibleConfiguration(ctx, subID, resourceGroup, serverName, mysqlConfigurationTLSVersion, nil, nil)
	if err != nil {
		return nil, err
	}

	if tlsVersion.Properties == nil {
		return nil, nil
	}

	return []AzureAsset{
		{
			Id:             pointers.Deref(tlsVersion.ID),
			Name:           pointers.Deref(tlsVersion.Name),
			ResourceGroup:  resourceGroup,
			SubscriptionId: subID,
			Location:       assetLocationGlobal,
			Properties: map[string]any{
				"source":       string(pointers.Deref(tlsVersion.Properties.Source)),
				"value":        strings.ToLower(pointers.Deref(tlsVersion.Properties.Value)),
				"dataType":     pointers.Deref(tlsVersion.Properties.DataType),
				"defaultValue": pointers.Deref(tlsVersion.Properties.DefaultValue),
			},
		},
	}, nil
}
