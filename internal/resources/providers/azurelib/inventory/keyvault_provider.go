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
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/maps"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type azureKeyVaultWrapper struct {
	AssetKeyVaultKeys       func(ctx context.Context, subscriptionID string, resourceGroupName string, vaultName string) ([]armkeyvault.KeysClientListResponse, error)
	AssetKeyVaultSecrets    func(ctx context.Context, subscriptionID string, resourceGroupName string, vaultName string) ([]armkeyvault.SecretsClientListResponse, error)
	AssetDiagnosticSettings func(ctx context.Context, resourceURI string, options *armmonitor.DiagnosticSettingsClientListOptions) ([]armmonitor.DiagnosticSettingsClientListResponse, error)
}

func defaultAzureKeyVaultWrapper(diagnosticSettingsClient *armmonitor.DiagnosticSettingsClient, credentials azcore.TokenCredential) *azureKeyVaultWrapper {
	return &azureKeyVaultWrapper{
		AssetKeyVaultKeys: func(ctx context.Context, subscriptionID string, resourceGroupName string, vaultName string) ([]armkeyvault.KeysClientListResponse, error) {
			client, err := armkeyvault.NewKeysClient(subscriptionID, credentials, nil)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, client.NewListPager(resourceGroupName, vaultName, nil))
		},
		AssetKeyVaultSecrets: func(ctx context.Context, subscriptionID string, resourceGroupName string, vaultName string) ([]armkeyvault.SecretsClientListResponse, error) {
			client, err := armkeyvault.NewSecretsClient(subscriptionID, credentials, nil)
			if err != nil {
				return nil, err
			}
			return readPager(ctx, client.NewListPager(resourceGroupName, vaultName, nil))
		},
		AssetDiagnosticSettings: func(ctx context.Context, resourceURI string, options *armmonitor.DiagnosticSettingsClientListOptions) ([]armmonitor.DiagnosticSettingsClientListResponse, error) {
			return readPager(ctx, diagnosticSettingsClient.NewListPager(resourceURI, options))
		},
	}
}

type KeyVaultProviderAPI interface {
	ListKeyVaultKeys(ctx context.Context, vault AzureAsset) ([]AzureAsset, error)
	ListKeyVaultSecrets(ctx context.Context, vault AzureAsset) ([]AzureAsset, error)
	ListKeyVaultDiagnosticSettings(ctx context.Context, vault AzureAsset) ([]AzureAsset, error)
}

func NewKeyVaultProvider(log *clog.Logger, diagnosticSettingsClient *armmonitor.DiagnosticSettingsClient, credentials azcore.TokenCredential) KeyVaultProviderAPI {
	return &keyVaultProvider{
		log:    log,
		client: defaultAzureKeyVaultWrapper(diagnosticSettingsClient, credentials),
	}
}

type keyVaultProvider struct {
	log    *clog.Logger
	client *azureKeyVaultWrapper
}

func (p *keyVaultProvider) ListKeyVaultDiagnosticSettings(ctx context.Context, vault AzureAsset) ([]AzureAsset, error) {
	p.log.Info("Listing Azure Vault Diagnostic Settings")

	responses, err := p.client.AssetDiagnosticSettings(ctx, vault.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving vault diagnostic settings: vaultId: %v, error: %w", vault.Id, err)
	}

	return lo.FlatMap(responses, func(res armmonitor.DiagnosticSettingsClientListResponse, _ int) []AzureAsset {
		return lo.FilterMap(res.Value, func(setting *armmonitor.DiagnosticSettingsResource, _ int) (AzureAsset, bool) {
			return p.transformDiagnosticSetting(setting, vault)
		})
	}), nil
}

func (p *keyVaultProvider) transformDiagnosticSetting(setting *armmonitor.DiagnosticSettingsResource, vault AzureAsset) (AzureAsset, bool) {
	if setting == nil {
		return AzureAsset{}, false
	}

	properties := map[string]any{}

	maps.AddIfNotNil(properties, "storageAccountId", setting.Properties.StorageAccountID)
	maps.AddIfSliceNotEmpty(properties, "logs", setting.Properties.Logs)

	if len(properties) == 0 {
		properties = nil
	}

	return AzureAsset{
		Id:             pointers.Deref(setting.ID),
		Name:           pointers.Deref(setting.Name),
		DisplayName:    "",
		Location:       "",
		ResourceGroup:  vault.ResourceGroup,
		SubscriptionId: vault.SubscriptionId,
		TenantId:       vault.TenantId,
		Type:           pointers.Deref(setting.Type),
		Properties:     properties,
	}, true
}

func (p *keyVaultProvider) ListKeyVaultKeys(ctx context.Context, vault AzureAsset) ([]AzureAsset, error) {
	p.log.Info("Listing Azure Vault Keys")

	responses, err := p.client.AssetKeyVaultKeys(ctx, vault.SubscriptionId, vault.ResourceGroup, vault.Name)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving vault keys: %w", err)
	}

	return lo.FlatMap(responses, func(res armkeyvault.KeysClientListResponse, _ int) []AzureAsset {
		return lo.FilterMap(res.Value, func(key *armkeyvault.Key, _ int) (AzureAsset, bool) {
			return p.transformKey(key, vault)
		})
	}), nil
}

func (p *keyVaultProvider) transformKey(key *armkeyvault.Key, vault AzureAsset) (AzureAsset, bool) {
	if key == nil {
		return AzureAsset{}, false
	}

	properties := map[string]any{}

	maps.AddIfMapNotEmpty(properties, "tags", key.Tags)
	maps.AddIfNotNil(properties, "attributes", key.Properties.Attributes)
	maps.AddIfNotNil(properties, "curveName", key.Properties.CurveName)
	maps.AddIfSliceNotEmpty(properties, "keyOps", key.Properties.KeyOps)
	maps.AddIfNotNil(properties, "keySize", key.Properties.KeySize)
	maps.AddIfNotNil(properties, "keyUri", key.Properties.KeyURI)
	maps.AddIfNotNil(properties, "keyUriWithVersion", key.Properties.KeyURIWithVersion)
	maps.AddIfNotNil(properties, "kty", key.Properties.Kty)
	maps.AddIfNotNil(properties, "releasePolicy", key.Properties.ReleasePolicy)
	maps.AddIfNotNil(properties, "rotationPolicy", key.Properties.RotationPolicy)
	if len(properties) == 0 {
		properties = nil
	}

	return AzureAsset{
		Id:             pointers.Deref(key.ID),
		Name:           pointers.Deref(key.Name),
		DisplayName:    "",
		Location:       pointers.Deref(key.Location),
		ResourceGroup:  vault.ResourceGroup,
		SubscriptionId: vault.SubscriptionId,
		TenantId:       vault.TenantId,
		Type:           pointers.Deref(key.Type),
		Properties:     properties,
	}, true
}

func (p *keyVaultProvider) ListKeyVaultSecrets(ctx context.Context, vault AzureAsset) ([]AzureAsset, error) {
	p.log.Info("Listing Azure Vault Secrets")

	responses, err := p.client.AssetKeyVaultSecrets(ctx, vault.SubscriptionId, vault.ResourceGroup, vault.Name)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving vault secrets: %w", err)
	}

	return lo.FlatMap(responses, func(res armkeyvault.SecretsClientListResponse, _ int) []AzureAsset {
		return lo.FilterMap(res.Value, func(secret *armkeyvault.Secret, _ int) (AzureAsset, bool) {
			return p.transformSecret(secret, vault)
		})
	}), nil
}

func (p *keyVaultProvider) transformSecret(secret *armkeyvault.Secret, vault AzureAsset) (AzureAsset, bool) {
	if secret == nil {
		return AzureAsset{}, false
	}

	properties := map[string]any{}

	maps.AddIfMapNotEmpty(properties, "tags", secret.Tags)
	maps.AddIfNotNil(properties, "attributes", secret.Properties.Attributes)
	maps.AddIfNotNil(properties, "contentType", secret.Properties.ContentType)
	maps.AddIfNotNil(properties, "secretUri", secret.Properties.SecretURI)
	maps.AddIfNotNil(properties, "secretUrlWithVersion", secret.Properties.SecretURIWithVersion)
	// secret.Properties.Value // do not use

	if len(properties) == 0 {
		properties = nil
	}

	return AzureAsset{
		Id:             pointers.Deref(secret.ID),
		Name:           pointers.Deref(secret.Name),
		DisplayName:    "",
		Location:       pointers.Deref(secret.Location),
		ResourceGroup:  vault.ResourceGroup,
		SubscriptionId: vault.SubscriptionId,
		TenantId:       vault.TenantId,
		Type:           pointers.Deref(secret.Type),
		Properties:     properties,
	}, true
}
