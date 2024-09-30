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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/utils/maps"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type storageAccountAzureClientWrapper struct {
	AssetDiagnosticSettings func(ctx context.Context, subID string, options *armmonitor.DiagnosticSettingsClientListOptions) ([]armmonitor.DiagnosticSettingsClientListResponse, error)
	AssetBlobServices       func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, resourceGroup, storageAccountName string, options *armstorage.BlobServicesClientListOptions) ([]armstorage.BlobServicesClientListResponse, error)
	AssetQueues             func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, resourceGroup, storageAccountName string, options *armstorage.QueueClientListOptions) ([]armstorage.QueueClientListResponse, error)
	AssetQueueServices      func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, resourceGroup, storageAccountName string, options *armstorage.QueueServicesClientListOptions) (armstorage.QueueServicesClientListResponse, error)
	AssetAccountStorage     func(ctx context.Context, subID string, clientOptions *arm.ClientOptions) ([]armstorage.AccountsClientListResponse, error)
}

type StorageAccountProviderAPI interface {
	ListDiagnosticSettingsAssetTypes(ctx context.Context, cycleMetadata cycle.Metadata, subscriptionIDs []string) ([]AzureAsset, error)
	ListStorageAccountBlobServices(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error)
	ListStorageAccountQueues(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error)
	ListStorageAccountQueueServices(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error)
	ListStorageAccountsBlobDiagnosticSettings(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error)
	ListStorageAccountsTableDiagnosticSettings(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error)
	ListStorageAccountsQueueDiagnosticSettings(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error)
	ListStorageAccounts(ctx context.Context, storageAccountsSubscriptionsIds []string) ([]AzureAsset, error)
}

type storageAccountProvider struct {
	client                  *storageAccountAzureClientWrapper
	log                     *logp.Logger
	diagnosticSettingsCache *cycle.Cache[[]AzureAsset]
}

func NewStorageAccountProvider(log *logp.Logger, diagnosticSettingsClient *armmonitor.DiagnosticSettingsClient, credentials azcore.TokenCredential) StorageAccountProviderAPI {
	// We wrap the client, so we can mock it in tests
	wrapper := &storageAccountAzureClientWrapper{
		AssetDiagnosticSettings: func(ctx context.Context, resourceURI string, options *armmonitor.DiagnosticSettingsClientListOptions) ([]armmonitor.DiagnosticSettingsClientListResponse, error) {
			pager := diagnosticSettingsClient.NewListPager(resourceURI, options)
			return readPager(ctx, pager)
		},
		AssetBlobServices: func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, resourceGroupName, storageAccountName string, options *armstorage.BlobServicesClientListOptions) ([]armstorage.BlobServicesClientListResponse, error) {
			cl, err := armstorage.NewBlobServicesClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListPager(resourceGroupName, storageAccountName, options))
		},
		AssetQueues: func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, resourceGroupName, storageAccountName string, options *armstorage.QueueClientListOptions) ([]armstorage.QueueClientListResponse, error) {
			cl, err := armstorage.NewQueueClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, cl.NewListPager(resourceGroupName, storageAccountName, options))
		},
		AssetQueueServices: func(ctx context.Context, subID string, clientOptions *arm.ClientOptions, resourceGroupName, storageAccountName string, options *armstorage.QueueServicesClientListOptions) (armstorage.QueueServicesClientListResponse, error) {
			cl, err := armstorage.NewQueueServicesClient(subID, credentials, clientOptions)
			if err != nil {
				return armstorage.QueueServicesClientListResponse{}, err
			}
			return cl.List(ctx, resourceGroupName, storageAccountName, options)
		},
		AssetAccountStorage: func(ctx context.Context, subID string, clientOptions *arm.ClientOptions) ([]armstorage.AccountsClientListResponse, error) {
			accountsClient, err := armstorage.NewAccountsClient(subID, credentials, clientOptions)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, accountsClient.NewListPager(&armstorage.AccountsClientListOptions{}))
		},
	}

	return &storageAccountProvider{
		log:                     log,
		client:                  wrapper,
		diagnosticSettingsCache: cycle.NewCache[[]AzureAsset](log),
	}
}

func (p *storageAccountProvider) ListStorageAccounts(ctx context.Context, storageAccountsSubscriptionsIds []string) ([]AzureAsset, error) {
	var assets []AzureAsset
	for _, saID := range storageAccountsSubscriptionsIds {
		res, err := p.client.AssetAccountStorage(ctx, saID, nil)
		if err != nil {
			p.log.Errorf("error while fetching storage accounts for subscriptionId: %s, error: %v", saID, err)
			continue
		}
		storageAccountsAssets, err := transformStorageAccounts(res, saID)
		if err != nil {
			p.log.Errorf("error while transforming storage for subscriptionId: %s, error: %v", saID, err)
			continue
		}
		assets = append(assets, storageAccountsAssets...)
	}
	return assets, nil
}

func transformStorageAccounts(accountPages []armstorage.AccountsClientListResponse, storageAccountId string) ([]AzureAsset, error) {
	var errs error
	return lo.FlatMap(accountPages, func(response armstorage.AccountsClientListResponse, _ int) []AzureAsset {
		return lo.Map(response.Value, func(item *armstorage.Account, _ int) AzureAsset {
			properties, err := maps.AsMapStringAny(item.Properties)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			return AzureAsset{
				Id:   pointers.Deref(item.ID),
				Name: pointers.Deref(item.Name),
				Type: strings.ToLower(pointers.Deref(item.Type)),
				// These are kept empty because:
				// 1. they are not used for evaluation
				// 2. only subscriptionId is required for the request issued by armstorage.NewAccountsClient
				ResourceGroup:  "",
				TenantId:       "",
				SubscriptionId: storageAccountId,
				Properties:     properties,
				Extension: map[string]any{
					ExtensionStorageAccountID:   pointers.Deref(item.ID),
					ExtensionStorageAccountName: pointers.Deref(item.Name),
				},
			}
		})
	}), errs
}

func (p *storageAccountProvider) ListStorageAccountBlobServices(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
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

func (p *storageAccountProvider) ListStorageAccountQueues(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	var assets []AzureAsset
	for _, sa := range storageAccounts {
		responses, err := p.client.AssetQueues(ctx, sa.SubscriptionId, nil, sa.ResourceGroup, sa.Name, nil)
		if err != nil {
			return nil, fmt.Errorf("error while fetching azure queues for storage accounts %s: %w", sa.Id, err)
		}

		queues, err := transformQueues(responses, sa)
		if err != nil {
			return nil, fmt.Errorf("error while transforming azure queues for storage accounts %s: %w", sa.Id, err)
		}

		assets = append(assets, queues...)
	}

	return assets, nil
}

func (p *storageAccountProvider) ListStorageAccountQueueServices(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	var assets []AzureAsset
	for _, sa := range storageAccounts {
		response, err := p.client.AssetQueueServices(ctx, sa.SubscriptionId, nil, sa.ResourceGroup, sa.Name, nil)
		if err != nil {
			return nil, fmt.Errorf("error while fetching azure queue services for storage accounts %s: %w", sa.Id, err)
		}

		var errs error
		for _, item := range response.Value {
			properties, err := maps.AsMapStringAny(item.QueueServiceProperties)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			assets = append(assets, AzureAsset{
				Id:             pointers.Deref(item.ID),
				Name:           pointers.Deref(item.Name),
				Type:           strings.ToLower(pointers.Deref(item.Type)),
				ResourceGroup:  sa.ResourceGroup,
				SubscriptionId: sa.SubscriptionId,
				TenantId:       sa.TenantId,
				Properties:     properties,
				Extension: map[string]any{
					ExtensionStorageAccountID:   sa.Id,
					ExtensionStorageAccountName: sa.Name,
				},
			})
		}
		if err != nil {
			return nil, fmt.Errorf("error while transforming azure queue services for storage accounts %s: %w", sa.Id, err)
		}
	}

	return assets, nil
}

func transformBlobServices(servicesPages []armstorage.BlobServicesClientListResponse, storageAccount AzureAsset) ([]AzureAsset, error) {
	var errs error
	return lo.FlatMap(servicesPages, func(response armstorage.BlobServicesClientListResponse, _ int) []AzureAsset {
		return lo.Map(response.Value, func(item *armstorage.BlobServiceProperties, _ int) AzureAsset {
			properties, err := maps.AsMapStringAny(item.BlobServiceProperties)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			return AzureAsset{
				Id:             pointers.Deref(item.ID),
				Name:           pointers.Deref(item.Name),
				Type:           strings.ToLower(pointers.Deref(item.Type)),
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

func transformQueues(pages []armstorage.QueueClientListResponse, storageAccount AzureAsset) ([]AzureAsset, error) {
	var errs error
	return lo.FlatMap(pages, func(response armstorage.QueueClientListResponse, _ int) []AzureAsset {
		return lo.Map(response.Value, func(item *armstorage.ListQueue, _ int) AzureAsset {
			properties, err := maps.AsMapStringAny(item.QueueProperties)
			if err != nil {
				errs = errors.Join(errs, err)
			}

			return AzureAsset{
				Id:             pointers.Deref(item.ID),
				Name:           pointers.Deref(item.Name),
				Type:           strings.ToLower(pointers.Deref(item.Type)),
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

func (p *storageAccountProvider) ListStorageAccountsBlobDiagnosticSettings(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	return p.serviceDiagnosticSettings(ctx, "blobServices/default", storageAccounts)
}

func (p *storageAccountProvider) ListStorageAccountsTableDiagnosticSettings(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	return p.serviceDiagnosticSettings(ctx, "tableServices/default", storageAccounts)
}

func (p *storageAccountProvider) ListStorageAccountsQueueDiagnosticSettings(ctx context.Context, storageAccounts []AzureAsset) ([]AzureAsset, error) {
	return p.serviceDiagnosticSettings(ctx, "queueServices/default", storageAccounts)
}

func (p *storageAccountProvider) ListDiagnosticSettingsAssetTypes(ctx context.Context, cycleMetadata cycle.Metadata, subscriptionIDs []string) ([]AzureAsset, error) {
	p.log.Info("Listing Azure Diagnostic Monitor Settings")

	return p.diagnosticSettingsCache.GetValue(ctx, cycleMetadata, func(ctx context.Context) ([]AzureAsset, error) {
		return p.getDiagnosticSettings(ctx, subscriptionIDs)
	})
}

func (p *storageAccountProvider) getDiagnosticSettings(ctx context.Context, subscriptionIDs []string) ([]AzureAsset, error) {
	var assets []AzureAsset

	for _, subID := range subscriptionIDs {
		responses, err := p.client.AssetDiagnosticSettings(ctx, fmt.Sprintf("/subscriptions/%s/", subID), nil)
		if err != nil {
			return nil, err
		}
		a, err := transformDiagnosticSettingsClientListResponses(responses, subID)
		if err != nil {
			return nil, err
		}
		assets = append(assets, a...)
	}

	return assets, nil
}

func (p *storageAccountProvider) serviceDiagnosticSettings(ctx context.Context, serviceIDPostfix string, storageAccounts []AzureAsset) ([]AzureAsset, error) {
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

func (p *storageAccountProvider) storageAccountServiceDiagnosticSettings(ctx context.Context, idPostfix string, storageAccount AzureAsset) ([]AzureAsset, error) {
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

func transformDiagnosticSettingsClientListResponses(response []armmonitor.DiagnosticSettingsClientListResponse, subID string) ([]AzureAsset, error) {
	var assets []AzureAsset

	for _, settingsCollection := range response {
		for _, v := range settingsCollection.Value {
			if v == nil {
				continue
			}
			a, err := transformDiagnosticSettingsResource(v, subID)
			if err != nil {
				return nil, fmt.Errorf("error parsing azure asset model: %w", err)
			}
			assets = append(assets, a)
		}
	}

	return assets, nil
}

func transformDiagnosticSettingsResource(v *armmonitor.DiagnosticSettingsResource, subID string) (AzureAsset, error) {
	properties, err := maps.AsMapStringAny(v.Properties)
	if err != nil {
		return AzureAsset{}, err
	}

	return AzureAsset{
		Id:             pointers.Deref(v.ID),
		Name:           pointers.Deref(v.Name),
		Location:       "global",
		Properties:     properties,
		ResourceGroup:  "",
		SubscriptionId: subID,
		TenantId:       "",
		Type:           pointers.Deref(v.Type),
	}, nil
}
