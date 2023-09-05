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
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/elastic-agent-libs/logp"
	mock "github.com/stretchr/testify/mock"
)

func TestConfigProvider_GetAzureClientConfig(t *testing.T) {
	type fields struct {
		AuthProvider AzureAuthProviderAPI
	}
	type args struct {
		cfg config.AzureConfig
		log *logp.Logger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *AzureFactoryConfig
		wantErr bool
	}{
		{
			name: "Should return a DefaultAzureCredential",
			fields: fields{
				AuthProvider: mockAzureAuthProvider(nil),
			},
			args: args{
				cfg: config.AzureConfig{},
				log: logp.NewLogger("Should return a DefaultAzureCredential"),
			},
			want: &AzureFactoryConfig{
				Credentials: &azidentity.DefaultAzureCredential{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ConfigProvider{
				AuthProvider: tt.fields.AuthProvider,
			}
			got, err := p.GetAzureClientConfig(tt.args.cfg, tt.args.log)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigProvider.GetAzureClientConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConfigProvider.GetAzureClientConfig() = %v, want %v", got, tt.want)
			}
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
		)
	} else {
		on.Return(nil, err)
	}
	return azureProviderAPI
}
