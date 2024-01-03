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
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/resources/utils/maps"
	cloudbeat_strings "github.com/elastic/cloudbeat/resources/utils/strings"
)

func (p *provider) ListStorageAccountBlobServices(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	var assets []AzureAsset

	for _, sa := range storageAccounts {
		responses, err := p.client.AssetBlobServices(ctx, sa.SubscriptionId, nil, sa.ResourceGroup, sa.Name, nil)
		if err != nil {
			return nil, fmt.Errorf("error while fetching azure blob services for storage accounts %s: %w", sa.Id, err)
		}

		blobServices, err := transformBlobServices(responses, sa)
		if err != nil {
			return nil, fmt.Errorf("error while transforming azure blob services for storage accounts %s: %w", sa.Id, err)
		}

		assets = append(assets, blobServices...)
	}

	return assets, nil
}

func transformBlobServices(servicesPages []armstorage.BlobServicesClientListResponse, storageAccount AzureAsset) (_ []AzureAsset, errs error) {
	return lo.FlatMap(servicesPages, func(response armstorage.BlobServicesClientListResponse, _ int) []AzureAsset {
		return lo.Map(response.Value, func(item *armstorage.BlobServiceProperties, _ int) AzureAsset {
			properties, err := maps.AsMapStringAny(item.BlobServiceProperties)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			return AzureAsset{
				Id:             cloudbeat_strings.Dereference(item.ID),
				Name:           cloudbeat_strings.Dereference(item.Name),
				Type:           strings.ToLower(cloudbeat_strings.Dereference(item.Type)),
				ResourceGroup:  storageAccount.ResourceGroup,
				SubscriptionId: storageAccount.SubscriptionId,
				TenantId:       storageAccount.TenantId,
				Properties:     properties,
				Extension: map[string]any{
					ExtensionStorageAccountID:   storageAccount.Id,
					ExtensionStorageAccountName: storageAccount.Name,
				},
			}
		})
	}), errs
}

func (p *provider) ListStorageAccountsBlobDiagnosticSettings(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	return p.serviceDiagnosticSettings(ctx, "blobServices/default", storageAccounts)
}

func (p *provider) ListStorageAccountsTableDiagnosticSettings(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	return p.serviceDiagnosticSettings(ctx, "tableServices/default", storageAccounts)
}

func (p *provider) ListStorageAccountsQueueDiagnosticSettings(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	return p.serviceDiagnosticSettings(ctx, "queueServices/default", storageAccounts)
}

func (p *provider) serviceDiagnosticSettings(ctx context.Context, serviceIDPostfix string, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	var assets []AzureAsset

	for _, sa := range storageAccounts {
		queueDiagSettings, err := p.storageAccountServiceDiagnosticSettings(ctx, serviceIDPostfix, sa)
		if err != nil {
			return nil, err
		}
		assets = append(assets, queueDiagSettings...)
	}

	return assets, nil
}

func (p *provider) storageAccountServiceDiagnosticSettings(ctx context.Context, idPostfix string, storageAccount AzureAsset) ([]AzureAsset, error) {
	res, err := p.client.AssetDiagnosticSettings(ctx, fmt.Sprintf("%s/%s", storageAccount.Id, idPostfix), nil)
	if err != nil {
		return nil, fmt.Errorf("error while fetching storage account service %s diagnostic settings: %w", idPostfix, err)
	}
	assets, err := transformDiagnosticSettingsClientListResponses(res, storageAccount.SubscriptionId)
	if err != nil {
		return nil, fmt.Errorf("error while transforming storage account service %s diagnostic settings: %w", idPostfix, err)
	}

	for i := range assets {
		(&assets[i]).AddExtension(ExtensionStorageAccountID, storageAccount.Id)
	}

	return assets, nil
}
