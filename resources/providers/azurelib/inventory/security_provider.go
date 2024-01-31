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
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/resources/utils/maps"
	"github.com/elastic/cloudbeat/resources/utils/pointers"
)

const listSecurityContactsURI = "/subscriptions/%s/providers/Microsoft.Security/securityContacts"

// azureSecurityContactsClient is needed till this issue -
// https://github.com/Azure/azure-sdk-for-go/issues/19740 is fixed.
// Implementation code is similar to armsecurity.ContactsClient but it attempts to decode -
// the List response to two possible structs.
type azureSecurityContactsClient struct {
	armClient *arm.Client
}

// ListSecurityContacts implements the Security Contacts List Azure REST API call.
// https://learn.microsoft.com/en-us/rest/api/defenderforcloud/security-contacts/list
func (c *azureSecurityContactsClient) ListSecurityContacts(ctx context.Context, subID string) (armsecurity.ContactsClientListResponse, error) {
	req, err := c.listSecurityContactsRequest(ctx, c.armClient.Endpoint(), subID)
	if err != nil {
		return armsecurity.ContactsClientListResponse{}, err
	}

	httpResp, err := c.armClient.Pipeline().Do(req)
	if err != nil {
		return armsecurity.ContactsClientListResponse{}, err
	}

	if !runtime.HasStatusCode(httpResp, http.StatusOK) {
		err = runtime.NewResponseError(httpResp)
		return armsecurity.ContactsClientListResponse{}, err
	}

	return c.listSecurityContactsResponseDecode(httpResp)
}

// listCreateRequest creates the List request.
func (*azureSecurityContactsClient) listSecurityContactsRequest(ctx context.Context, endpoint, subID string) (*policy.Request, error) {
	if subID == "" {
		return nil, errors.New("parameter subscription id cannot be empty")
	}

	urlPath := fmt.Sprintf(listSecurityContactsURI, url.PathEscape(subID))

	req, err := runtime.NewRequest(ctx, http.MethodGet, runtime.JoinPaths(endpoint, urlPath))
	if err != nil {
		return nil, err
	}

	reqQP := req.Raw().URL.Query()
	reqQP.Set("api-version", "2020-01-01-preview")
	req.Raw().URL.RawQuery = reqQP.Encode()
	req.Raw().Header.Set("Accept", "application/json")

	return req, nil
}

func (*azureSecurityContactsClient) listSecurityContactsResponseDecode(resp *http.Response) (armsecurity.ContactsClientListResponse, error) {
	result := armsecurity.ContactsClientListResponse{}
	if err := runtime.UnmarshalAsJSON(resp, &result.ContactList.Value); err != nil {
		// fallback attempt to unmarshal
		if secondErr := runtime.UnmarshalAsJSON(resp, &result); secondErr != nil {
			return armsecurity.ContactsClientListResponse{}, err
		}
	}
	return result, nil
}

type securityContactsClient interface {
	ListSecurityContacts(ctx context.Context, subID string) (armsecurity.ContactsClientListResponse, error)
}

type SecurityContactsProviderAPI interface {
	ListSecurityContacts(ctx context.Context, subscriptionID string) ([]AzureAsset, error)
}

type securityContactsProvider struct {
	client securityContactsClient
	log    *logp.Logger //nolint:unused
}

func NewSecurityContacts(log *logp.Logger, armClient *arm.Client) SecurityContactsProviderAPI {
	return &securityContactsProvider{
		log: log,
		client: &azureSecurityContactsClient{
			armClient: armClient,
		},
	}
}

func (p *securityContactsProvider) ListSecurityContacts(ctx context.Context, subscriptionID string) ([]AzureAsset, error) {
	res, err := p.client.ListSecurityContacts(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("error while fetching security contacts: %w", err)
	}

	return lo.FilterMap(res.Value, func(contract *armsecurity.Contact, _ int) (AzureAsset, bool) {
		if contract == nil {
			return AzureAsset{}, false
		}
		return p.transformSecurityContract(contract, subscriptionID), true
	}), nil
}

func (p *securityContactsProvider) transformSecurityContract(contact *armsecurity.Contact, subscriptionID string) AzureAsset {
	properties := map[string]any{}

	maps.AddIfNotNil(properties, "alertNotifications", contact.Properties.AlertNotifications)
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
