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
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConfigProvider_GetAzureClientConfig(t *testing.T) {
	tests := []struct {
		name         string
		want         *AzureFactoryConfig
		wantErr      bool
		authProvider *MockAzureAuthProviderAPI
	}{
		{
			name:         "Should return a DefaultAzureCredential",
			authProvider: mockAzureAuthProvider(nil),
			want: &AzureFactoryConfig{
				Credentials: &azidentity.DefaultAzureCredential{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.authProvider.AssertExpectations(t)

			p := &ConfigProvider{
				AuthProvider: tt.authProvider,
			}
			got, err := p.GetAzureClientConfig()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func mockAzureAuthProvider(err error) *MockAzureAuthProviderAPI {
	azureProviderAPI := &MockAzureAuthProviderAPI{}
	on := azureProviderAPI.EXPECT().FindDefaultCredentials(mock.Anything)
	if err == nil {
		on.Return(
			&azidentity.DefaultAzureCredential{},
			nil,
		).Once()
	} else {
		on.Return(nil, err).Once()
	}
	return azureProviderAPI
}
