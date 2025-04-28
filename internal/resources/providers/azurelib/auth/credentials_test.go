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

package auth

import (
	"errors"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/config"
)

func TestConfigProvider_GetAzureClientConfig(t *testing.T) {
	tests := []struct {
		name               string
		config             config.AzureConfig
		authProviderInitFn func(*MockAzureAuthProviderAPI)
		want               *AzureFactoryConfig
		wantErr            bool
	}{
		{
			name:               "Should return a DefaultAzureCredential",
			config:             config.AzureConfig{},
			authProviderInitFn: initDefaultCredentialsMock(nil),
			want: &AzureFactoryConfig{
				Credentials: &azidentity.DefaultAzureCredential{},
			},
		},
		{
			name: "Should return a DefaultAzureCredential using managed_identity type",
			config: config.AzureConfig{
				Credentials: config.AzureClientOpt{
					ClientCredentialsType: config.AzureClientCredentialsTypeManagedIdentity,
				},
			},
			authProviderInitFn: initDefaultCredentialsMock(nil),
			want: &AzureFactoryConfig{
				Credentials: &azidentity.DefaultAzureCredential{},
			},
		},
		{
			name: "Should return a DefaultAzureCredential using manual type",
			config: config.AzureConfig{
				Credentials: config.AzureClientOpt{
					ClientCredentialsType: config.AzureClientCredentialsTypeManual,
				},
			},
			authProviderInitFn: initDefaultCredentialsMock(nil),
			want: &AzureFactoryConfig{
				Credentials: &azidentity.DefaultAzureCredential{},
			},
		},
		{
			name: "Should return a DefaultAzureCredential using arm_template type",
			config: config.AzureConfig{
				Credentials: config.AzureClientOpt{
					ClientCredentialsType: config.AzureClientCredentialsTypeARMTemplate,
				},
			},
			authProviderInitFn: initDefaultCredentialsMock(nil),
			want: &AzureFactoryConfig{
				Credentials: &azidentity.DefaultAzureCredential{},
			},
		},
		{
			name: "Should return a error on unknown client credentials type",
			config: config.AzureConfig{
				Credentials: config.AzureClientOpt{
					ClientCredentialsType: "unknown",
				},
			},
			authProviderInitFn: func(_ *MockAzureAuthProviderAPI) {},
			want:               nil,
			wantErr:            true,
		},
		{
			name: "Should return a ClientSecretCredential",
			config: config.AzureConfig{
				Credentials: config.AzureClientOpt{
					ClientCredentialsType: config.AzureClientCredentialsTypeSecret,
					TenantID:              "tenant_a",
					ClientID:              "client_id",
					ClientSecret:          "secret",
				},
			},
			authProviderInitFn: func(m *MockAzureAuthProviderAPI) {
				m.EXPECT().
					FindClientSecretCredentials("tenant_a", "client_id", "secret", mock.Anything).
					Return(&azidentity.ClientSecretCredential{}, nil).
					Once()
			},
			want: &AzureFactoryConfig{
				Credentials: &azidentity.ClientSecretCredential{},
			},
			wantErr: false,
		},
		{
			name: "Should return a ClientCertificateCredential",
			config: config.AzureConfig{
				Credentials: config.AzureClientOpt{
					ClientCredentialsType:     config.AzureClientCredentialsTypeCertificate,
					TenantID:                  "tenant_a",
					ClientID:                  "client_id",
					ClientCertificatePath:     "/path/cert",
					ClientCertificatePassword: "password",
				},
			},
			authProviderInitFn: func(m *MockAzureAuthProviderAPI) {
				m.EXPECT().
					FindCertificateCredential("tenant_a", "client_id", "/path/cert", "password", mock.Anything).
					Return(&azidentity.ClientCertificateCredential{}, nil).
					Once()
			},
			want: &AzureFactoryConfig{
				Credentials: &azidentity.ClientCertificateCredential{},
			},
			wantErr: false,
		},
		{
			name: "Should return an error on credentials error",
			config: config.AzureConfig{
				Credentials: config.AzureClientOpt{
					ClientCredentialsType: config.AzureClientCredentialsTypeManagedIdentity,
				},
			},
			authProviderInitFn: initDefaultCredentialsMock(errMockAzure),
			want:               nil,
			wantErr:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			azureProviderAPI := &MockAzureAuthProviderAPI{}
			tt.authProviderInitFn(azureProviderAPI)
			defer azureProviderAPI.AssertExpectations(t)

			p := &ConfigProvider{
				AuthProvider: azureProviderAPI,
			}
			got, err := p.GetAzureClientConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func initDefaultCredentialsMock(err error) func(*MockAzureAuthProviderAPI) {
	return func(azureProviderAPI *MockAzureAuthProviderAPI) {
		on := azureProviderAPI.EXPECT().FindDefaultCredentials(mock.Anything)
		if err == nil {
			on.Return(
				&azidentity.DefaultAzureCredential{},
				nil,
			).Once()
		} else {
			on.Return(nil, err).Once()
		}
	}
}

var errMockAzure = errors.New("mock azure error")
