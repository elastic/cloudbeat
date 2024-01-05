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

	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/providers/azurelib"
	"github.com/elastic/cloudbeat/resources/providers/azurelib/inventory"
)

type sqlServerEnricher struct {
	provider azurelib.ProviderAPI
}

func (s sqlServerEnricher) Enrich(ctx context.Context, _ cycle.Metadata, assets []inventory.AzureAsset) error {
	var errs error

	enrichFn := []func(context.Context, *inventory.AzureAsset) error{
		s.enrichSQLEncryptionProtector,
		s.enrichSQLBlobAuditPolicy,
		s.enrichTransparentDataEncryption,
	}

	for i, a := range assets {
		if a.Type != inventory.SQLServersAssetType {
			continue
		}

		for _, fn := range enrichFn {
			if err := fn(ctx, &a); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		assets[i] = a
	}

	return errs
}

func (s sqlServerEnricher) enrichSQLEncryptionProtector(ctx context.Context, a *inventory.AzureAsset) error {
	encryptProtectors, err := s.provider.ListSQLEncryptionProtector(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
	if err != nil {
		return err
	}

	if len(encryptProtectors) == 0 {
		return nil
	}

	props := make([]map[string]any, 0, len(encryptProtectors))
	for _, ep := range encryptProtectors {
		props = append(props, ep.Properties)
	}

	a.AddExtension(inventory.ExtensionSQLEncryptionProtectors, props)
	return nil
}

func (s sqlServerEnricher) enrichSQLBlobAuditPolicy(ctx context.Context, a *inventory.AzureAsset) error {
	policy, err := s.provider.GetSQLBlobAuditingPolicies(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
	if err != nil {
		return err
	}

	if len(policy) == 0 {
		return nil
	}

	a.AddExtension(inventory.ExtensionSQLBlobAuditPolicy, policy[0].Properties)

	return nil
}

func (s sqlServerEnricher) enrichTransparentDataEncryption(ctx context.Context, a *inventory.AzureAsset) error {
	tdes, err := s.provider.ListSQLTransparentDataEncryptions(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
	if err != nil {
		return err
	}

	if len(tdes) == 0 {
		return nil
	}

	props := make([]map[string]any, 0, len(tdes))
	for _, tde := range tdes {
		props = append(props, tde.Properties)
	}

	a.AddExtension(inventory.ExtensionSQLTransparentDataEncryptions, props)
	return nil
}
