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
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/maps"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type securityClientWrapper interface {
	ListSecurityContacts(ctx context.Context, subID string) ([]armsecurity.ContactsClientListResponse, error)
	ListAutoProvisioningSettings(ctx context.Context, subID string) ([]armsecurity.AutoProvisioningSettingsClientListResponse, error)
}

type SecurityContactsProviderAPI interface {
	ListSecurityContacts(ctx context.Context, subscriptionID string) ([]AzureAsset, error)
	ListAutoProvisioningSettings(ctx context.Context, subscriptionID string) ([]AzureAsset, error)
}

type defaultSecurityClientWrapper struct {
	credential azcore.TokenCredential
}

func (w *defaultSecurityClientWrapper) ListSecurityContacts(ctx context.Context, subID string) ([]armsecurity.ContactsClientListResponse, error) {
	cl, err := armsecurity.NewContactsClient(subID, w.credential, nil)
	if err != nil {
		return nil, err
	}

	return readPager(ctx, cl.NewListPager(nil))
}

func (w *defaultSecurityClientWrapper) ListAutoProvisioningSettings(ctx context.Context, subID string) ([]armsecurity.AutoProvisioningSettingsClientListResponse, error) {
	cl, err := armsecurity.NewAutoProvisioningSettingsClient(subID, w.credential, nil)
	if err != nil {
		return nil, err
	}
	return readPager(ctx, cl.NewListPager(nil))
}

type securityContactsProvider struct {
	client securityClientWrapper
	log    *clog.Logger //nolint:unused
}

func NewSecurityContacts(log *clog.Logger, credential azcore.TokenCredential) SecurityContactsProviderAPI {
	return &securityContactsProvider{
		log: log,
		client: &defaultSecurityClientWrapper{
			credential: credential,
		},
	}
}

func (p *securityContactsProvider) ListSecurityContacts(ctx context.Context, subscriptionID string) ([]AzureAsset, error) {
	res, err := p.client.ListSecurityContacts(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("error while fetching security contacts: %w", err)
	}

	return lo.Flatten(
		lo.Map(res, func(listResponse armsecurity.ContactsClientListResponse, _ int) []AzureAsset {
			return lo.FilterMap(listResponse.ContactList.Value, func(contact *armsecurity.Contact, _ int) (AzureAsset, bool) {
				if contact == nil {
					return AzureAsset{}, false
				}
				return p.transformSecurityContract(contact, subscriptionID), true
			})
		}),
	), nil
}

func (p *securityContactsProvider) transformSecurityContract(contact *armsecurity.Contact, subscriptionID string) AzureAsset {
	properties := map[string]any{}

	maps.AddIfSliceNotEmpty(properties, "notificationsSources", contact.Properties.NotificationsSources)
	maps.AddIfNotNil(properties, "emails", contact.Properties.Emails)
	maps.AddIfNotNil(properties, "notificationsByRole", contact.Properties.NotificationsByRole)
	maps.AddIfNotNil(properties, "phone", contact.Properties.Phone)

	return AzureAsset{
		Id:             pointers.Deref(contact.ID),
		Name:           pointers.Deref(contact.Name),
		Type:           strings.ToLower(pointers.Deref(contact.Type)),
		SubscriptionId: subscriptionID,
		Properties:     properties,
	}
}

func (p *securityContactsProvider) ListAutoProvisioningSettings(ctx context.Context, subscriptionID string) ([]AzureAsset, error) {
	res, err := p.client.ListAutoProvisioningSettings(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("error while fetching security contacts: %w", err)
	}

	return lo.FlatMap(res, func(response armsecurity.AutoProvisioningSettingsClientListResponse, _ int) []AzureAsset {
		return lo.FilterMap(response.Value, func(contract *armsecurity.AutoProvisioningSetting, _ int) (AzureAsset, bool) {
			if contract == nil {
				return AzureAsset{}, false
			}
			return p.transformAutoProvisioningSetting(contract, subscriptionID), true
		})
	}), nil
}

func (p *securityContactsProvider) transformAutoProvisioningSetting(settings *armsecurity.AutoProvisioningSetting, subscriptionID string) AzureAsset {
	properties := map[string]any{}

	if settings.Properties != nil {
		maps.AddIfNotNil(properties, "autoProvision", settings.Properties.AutoProvision)
	}

	return AzureAsset{
		Id:             pointers.Deref(settings.ID),
		Name:           pointers.Deref(settings.Name),
		Type:           strings.ToLower(pointers.Deref(settings.Type)), // Microsoft.Security/autoProvisioningSettings
		SubscriptionId: subscriptionID,
		Properties:     properties,
	}
}
