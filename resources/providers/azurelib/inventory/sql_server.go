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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql"
	"github.com/samber/lo"
)

func (p *provider) ListSQLEncryptionProtector(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]AzureAsset, error) {
	encryptProtectors, err := p.client.AssetSQLEncryptionProtector(ctx, subID, resourceGroup, sqlServerName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("problem on listing sql encryption protectors (%w)", err)
	}

	capacity := lo.Reduce(encryptProtectors, func(acc int, i armsql.EncryptionProtectorsClientListByServerResponse, _ int) int {
		return acc + len(i.Value)
	}, 0)

	if capacity == 0 {
		return nil, nil
	}

	assets := make([]AzureAsset, 0, capacity)
	for _, epWrapper := range encryptProtectors {
		for _, ep := range epWrapper.Value {
			if ep == nil || ep.Properties == nil {
				continue
			}

			assets = append(assets, convertEncryptionProtector(ep, resourceGroup, subID))
		}
	}

	return assets, nil
}

func (p *provider) GetSQLBlobAuditingPolicies(ctx context.Context, subID, resourceGroup, sqlServerName string) ([]AzureAsset, error) {
	policy, err := p.client.AssetSQLBlobAuditingPolicies(ctx, subID, resourceGroup, sqlServerName, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("problem on getting sql blob auditing policies (%w)", err)
	}

	if policy.Properties == nil {
		return nil, nil
	}

	return []AzureAsset{
		{
			Id:       deref(policy.ID),
			Name:     deref(policy.Name),
			Location: "global",
			Properties: map[string]any{
				"state":                        string(deref(policy.Properties.State)),
				"isAzureMonitorTargetEnabled":  deref(policy.Properties.IsAzureMonitorTargetEnabled),
				"isDevopsAuditEnabled":         deref(policy.Properties.IsDevopsAuditEnabled),
				"isManagedIdentityInUse":       deref(policy.Properties.IsManagedIdentityInUse),
				"isStorageSecondaryKeyInUse":   deref(policy.Properties.IsStorageSecondaryKeyInUse),
				"queueDelayMs":                 deref(policy.Properties.QueueDelayMs),
				"retentionDays":                deref(policy.Properties.RetentionDays),
				"storageAccountAccessKey":      deref(policy.Properties.StorageAccountAccessKey),
				"storageAccountSubscriptionID": deref(policy.Properties.StorageAccountSubscriptionID),
				"storageEndpoint":              deref(policy.Properties.StorageEndpoint),

				"auditActionsAndGroups": lo.Map(policy.Properties.AuditActionsAndGroups, func(s *string, _ int) string {
					return deref(s)
				}),
			},
			ResourceGroup:  resourceGroup,
			SubscriptionId: subID,
			TenantId:       "",
			Type:           deref(policy.Type),
		},
	}, nil
}

func convertEncryptionProtector(ep *armsql.EncryptionProtector, resourceGroup string, subID string) AzureAsset {
	return AzureAsset{
		Id:       deref(ep.ID),
		Name:     deref(ep.Name),
		Location: deref(ep.Location),
		Properties: map[string]any{
			"kind":                deref(ep.Kind),
			"serverKeyType":       string(deref(ep.Properties.ServerKeyType)),
			"autoRotationEnabled": deref(ep.Properties.AutoRotationEnabled),
			"serverKeyName":       deref(ep.Properties.ServerKeyName),
			"subregion":           deref(ep.Properties.Subregion),
			"thumbprint":          deref(ep.Properties.Thumbprint),
			"uri":                 deref(ep.Properties.URI),
		},
		ResourceGroup:  resourceGroup,
		SubscriptionId: subID,
		TenantId:       "",
		Type:           deref(ep.Type),
	}
}

func deref[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}
