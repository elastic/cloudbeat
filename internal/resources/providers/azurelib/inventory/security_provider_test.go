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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestListSecurityContacts(t *testing.T) {
	notificationsByRole := func(roles []string, state string) *armsecurity.ContactPropertiesNotificationsByRole {
		n := &armsecurity.ContactPropertiesNotificationsByRole{}
		n.State = (*armsecurity.State)(&state)
		n.Roles = make([]*armsecurity.SecurityContactRole, 0, len(roles))
		for _, r := range roles {
			n.Roles = append(n.Roles, (*armsecurity.SecurityContactRole)(&r))
		}
		return n
	}

	response := func(c ...*armsecurity.Contact) []armsecurity.ContactsClientListResponse {
		return []armsecurity.ContactsClientListResponse{
			{
				ContactList: armsecurity.ContactList{
					Value: c,
				},
			},
		}
	}

	tests := map[string]struct {
		inputSubID      string
		mockReturn      []armsecurity.ContactsClientListResponse
		mockReturnError error
		expected        []AzureAsset
		expectError     bool
	}{
		"nil": {
			inputSubID:      "sub1",
			mockReturn:      response(),
			mockReturnError: nil,
			expected:        []AzureAsset{},
			expectError:     false,
		},
		"single contact": {
			inputSubID: "sub1",
			mockReturn: response(
				&armsecurity.Contact{
					ID:   to.Ptr("id1"),
					Name: to.Ptr("default"),
					Type: to.Ptr("Microsoft.Security/securityContacts"),
					Properties: &armsecurity.ContactProperties{
						NotificationsByRole: notificationsByRole([]string{"Owner"}, "On"),
					},
				},
			),
			mockReturnError: nil,
			expected: []AzureAsset{
				{
					Id:             "id1",
					Name:           "default",
					SubscriptionId: "sub1",
					Type:           "microsoft.security/securitycontacts",
					Properties: map[string]any{
						"notificationsByRole": notificationsByRole([]string{"Owner"}, "On"),
					},
				},
			},
			expectError: false,
		},
		"multi contact": {
			inputSubID: "sub1",
			mockReturn: response(
				&armsecurity.Contact{
					ID:   to.Ptr("id1"),
					Name: to.Ptr("default"),
					Type: to.Ptr("Microsoft.Security/securityContacts"),
					Properties: &armsecurity.ContactProperties{
						NotificationsByRole: notificationsByRole([]string{"Owner"}, "On"),
					},
				},
				&armsecurity.Contact{
					ID:   to.Ptr("id2"),
					Name: to.Ptr("non-default"),
					Type: to.Ptr("Microsoft.Security/securityContacts"),
					Properties: &armsecurity.ContactProperties{
						NotificationsByRole: notificationsByRole([]string{"Owner", "Admin"}, "On"),
					},
				},
			),
			mockReturnError: nil,
			expected: []AzureAsset{
				{
					Id:             "id1",
					Name:           "default",
					SubscriptionId: "sub1",
					Type:           "microsoft.security/securitycontacts",
					Properties: map[string]any{
						"notificationsByRole": notificationsByRole([]string{"Owner"}, "On"),
					},
				},
				{
					Id:             "id2",
					Name:           "non-default",
					SubscriptionId: "sub1",
					Type:           "microsoft.security/securitycontacts",
					Properties: map[string]any{
						"notificationsByRole": notificationsByRole([]string{"Owner", "Admin"}, "On"),
					},
				},
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			log := testhelper.NewLogger(t)

			mockAzure := newMockSecurityClientWrapper(t)
			mockAzure.EXPECT().ListSecurityContacts(mock.Anything, tc.inputSubID).Return(tc.mockReturn, tc.mockReturnError)

			prov := securityContactsProvider{
				log:    log,
				client: mockAzure,
			}

			got, gotErr := prov.ListSecurityContacts(context.Background(), tc.inputSubID)
			if tc.expectError {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			require.Equal(t, tc.expected, got)
		})
	}
}

func TestListAutoProvisioningSettings(t *testing.T) {
	resourceType := "Microsoft.Security/autoProvisioningSettings"

	responses := func(r ...armsecurity.AutoProvisioningSettingsClientListResponse) []armsecurity.AutoProvisioningSettingsClientListResponse {
		return r
	}

	response := func(s ...*armsecurity.AutoProvisioningSetting) armsecurity.AutoProvisioningSettingsClientListResponse {
		return armsecurity.AutoProvisioningSettingsClientListResponse{
			AutoProvisioningSettingList: armsecurity.AutoProvisioningSettingList{
				Value: s,
			},
		}
	}

	settings := func(id, name string, autoProvision *string) *armsecurity.AutoProvisioningSetting {
		a := &armsecurity.AutoProvisioningSetting{
			ID:   to.Ptr(id),
			Name: to.Ptr(name),
			Type: to.Ptr(resourceType),
		}

		if autoProvision != nil {
			a.Properties = &armsecurity.AutoProvisioningSettingProperties{
				AutoProvision: (*armsecurity.AutoProvision)(autoProvision),
			}
		}

		return a
	}

	tests := map[string]struct {
		inputSubID      string
		mockReturn      []armsecurity.AutoProvisioningSettingsClientListResponse
		mockReturnError error
		expected        []AzureAsset
		expectError     bool
	}{
		"nil": {
			inputSubID:      "sub1",
			mockReturn:      responses(),
			mockReturnError: nil,
			expected:        []AzureAsset{},
			expectError:     false,
		},
		"single contact": {
			inputSubID: "sub1",
			mockReturn: responses(
				response(settings("id1", "default", to.Ptr("On"))),
			),
			mockReturnError: nil,
			expected: []AzureAsset{
				{
					Id:             "id1",
					Name:           "default",
					SubscriptionId: "sub1",
					Type:           "microsoft.security/autoprovisioningsettings",
					Properties: map[string]any{
						"autoProvision": (*armsecurity.AutoProvision)(to.Ptr("On")),
					},
				},
			},
			expectError: false,
		},
		"multi contact": {
			inputSubID: "sub1",
			mockReturn: responses(
				response(
					settings("id1", "default", to.Ptr("On")),
					settings("id2", "non-default", to.Ptr("Off")),
				),
			),
			mockReturnError: nil,
			expected: []AzureAsset{
				{
					Id:             "id1",
					Name:           "default",
					SubscriptionId: "sub1",
					Type:           "microsoft.security/autoprovisioningsettings",
					Properties: map[string]any{
						"autoProvision": (*armsecurity.AutoProvision)(to.Ptr("On")),
					},
				},
				{
					Id:             "id2",
					Name:           "non-default",
					SubscriptionId: "sub1",
					Type:           "microsoft.security/autoprovisioningsettings",
					Properties: map[string]any{
						"autoProvision": (*armsecurity.AutoProvision)(to.Ptr("Off")),
					},
				},
			},
			expectError: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			log := testhelper.NewLogger(t)

			mockAzure := newMockSecurityClientWrapper(t)
			mockAzure.EXPECT().ListAutoProvisioningSettings(mock.Anything, tc.inputSubID).Return(tc.mockReturn, tc.mockReturnError)

			prov := securityContactsProvider{
				log:    log,
				client: mockAzure,
			}

			got, gotErr := prov.ListAutoProvisioningSettings(context.Background(), tc.inputSubID)
			if tc.expectError {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			require.Equal(t, tc.expected, got)
		})
	}
}
