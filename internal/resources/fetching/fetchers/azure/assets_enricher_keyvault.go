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

package fetchers

import (
	"context"
	"errors"

	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/internal/resources/providers/azurelib/inventory"
)

type keyVaultEnricher struct {
	provider azurelib.ProviderAPI
}

func (e keyVaultEnricher) Enrich(ctx context.Context, _ cycle.Metadata, assets []inventory.AzureAsset) error {
	singleAssetEnrichers := []func(context.Context, *inventory.AzureAsset) error{
		e.enrichKeyVaultDiagnosticSettings,
		e.enrichKeyVaultWithKeys,
		e.enrichKeyVaultWithSecrets,
	}

	var errs []error
	for i, a := range assets {
		if a.Type != inventory.VaultAssetType {
			continue
		}

		for _, fn := range singleAssetEnrichers {
			if err := fn(ctx, &a); err != nil {
				errs = append(errs, err)
			}
		}

		assets[i] = a
	}

	return errors.Join(errs...)
}

func (e keyVaultEnricher) enrichKeyVaultDiagnosticSettings(ctx context.Context, a *inventory.AzureAsset) error {
	diagSettings, err := e.provider.ListKeyVaultDiagnosticSettings(ctx, *a)
	if err != nil {
		return err
	}

	if len(diagSettings) == 0 {
		return nil
	}

	a.AddExtension(inventory.ExtensionKeyVaultDiagnosticSettings, diagSettings)
	return nil
}

func (e keyVaultEnricher) enrichKeyVaultWithKeys(ctx context.Context, a *inventory.AzureAsset) error {
	keys, err := e.provider.ListKeyVaultKeys(ctx, *a)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	a.AddExtension(inventory.ExtensionKeyVaultKeys, keys)
	return nil
}

func (e keyVaultEnricher) enrichKeyVaultWithSecrets(ctx context.Context, a *inventory.AzureAsset) error {
	keys, err := e.provider.ListKeyVaultSecrets(ctx, *a)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	a.AddExtension(inventory.ExtensionKeyVaultSecrets, keys)
	return nil
}
