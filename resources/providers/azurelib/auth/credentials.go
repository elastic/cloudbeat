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
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/config"
)

type AzureFactoryConfig struct {
	// TODO: Add other credentials
	Credentials *azidentity.DefaultAzureCredential
}

type ConfigProviderAPI interface {
	GetAzureClientConfig(cfg config.AzureConfig, log *logp.Logger) (*AzureFactoryConfig, error)
}

type AzureAuthProviderAPI interface {
	FindDefaultCredentials(options *azidentity.DefaultAzureCredentialOptions) (*azidentity.DefaultAzureCredential, error)
	// FindEnvironmentCredential(options *azidentity.EnvironmentCredentialOptions) (*azidentity.EnvironmentCredential, error)
}

type ConfigProvider struct {
	AuthProvider AzureAuthProviderAPI
}

func (p *ConfigProvider) GetAzureClientConfig(cfg config.AzureConfig, log *logp.Logger) (*AzureFactoryConfig, error) {
	// if cfg.ClientId != "" {
	// 	return p.getCustomCredentialsConfig(cfg, log)
	// }
	return p.getDefaultCredentialsConfig(log)
}

func (p *ConfigProvider) getDefaultCredentialsConfig(log *logp.Logger) (*AzureFactoryConfig, error) {
	log.Info("getDefaultCredentialsConfig")

	creds, err := p.AuthProvider.FindDefaultCredentials(nil)
	if err != nil {
		return nil, fmt.Errorf("getDefaultCredentialsConfig failed to get default credentials: %v", err)
	}

	return &AzureFactoryConfig{
		Credentials: creds,
	}, nil
}

// func (p *ConfigProvider) getCustomCredentialsConfig(cfg config.AzureConfig, log *logp.Logger) (*AzureFactoryConfig, error) {
// 	log.Info("getCustomCredentialsConfig")

// 	creds, err := p.AuthProvider.FindEnvironmentCredential(&azidentity.EnvironmentCredentialOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("getCustomCredentialsConfig failed to get default credentials: %v", err)
// 	}

// 	return &AzureFactoryConfig{
// 		Credentials: creds,
// 	}, nil
// }
