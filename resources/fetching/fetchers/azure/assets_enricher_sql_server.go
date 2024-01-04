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
	for i, a := range assets {
		if a.Type != inventory.SQLServersAssetType {
			continue
		}

		props, err := s.fetchEncryptionProtectorProps(ctx, a)
		if err != nil {
			errs = errors.Join(errs, err)
		}

		if len(props) == 0 {
			continue
		}

		a.AddExtension(inventory.ExtensionEncryptionProtectors, props)
		assets[i] = a
	}

	return errs
}

func (s sqlServerEnricher) fetchEncryptionProtectorProps(ctx context.Context, a inventory.AzureAsset) ([]map[string]any, error) {
	encryptProtectors, err := s.provider.ListSQLEncryptionProtector(ctx, a.SubscriptionId, a.ResourceGroup, a.Name)
	if err != nil {
		return nil, err
	}

	protectorsProps := make([]map[string]any, 0, len(encryptProtectors))
	for _, ep := range encryptProtectors {
		protectorsProps = append(protectorsProps, ep.Properties)
	}
	return protectorsProps, nil
}
