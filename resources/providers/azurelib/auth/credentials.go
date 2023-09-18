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
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type AzureFactoryConfig struct {
	Credentials *azidentity.DefaultAzureCredential
}

type ConfigProviderAPI interface {
	GetAzureClientConfig() (*AzureFactoryConfig, error)
}

type ConfigProvider struct {
	AuthProvider AzureAuthProviderAPI
}

func (p *ConfigProvider) GetAzureClientConfig() (*AzureFactoryConfig, error) {
	return p.getDefaultCredentialsConfig()
}

func (p *ConfigProvider) getDefaultCredentialsConfig() (*AzureFactoryConfig, error) {
	creds, err := p.AuthProvider.FindDefaultCredentials(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get default credentials: %w", err)
	}

	return &AzureFactoryConfig{
		Credentials: creds,
	}, nil
}
