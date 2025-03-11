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
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	azlog "github.com/Azure/azure-sdk-for-go/sdk/azcore/log"

	// print log output to stdout
	"github.com/elastic/cloudbeat/internal/config"
)

type AzureFactoryConfig struct {
	Credentials azcore.TokenCredential
}

type ConfigProviderAPI interface {
	GetAzureClientConfig(cfg config.AzureConfig) (*AzureFactoryConfig, error)
}

type ConfigProvider struct {
	AuthProvider AzureAuthProviderAPI
}

func (p *ConfigProvider) GetAzureClientConfig(cfg config.AzureConfig) (*AzureFactoryConfig, error) {
	azlog.SetListener(func(event azlog.Event, s string) {
		fmt.Println(s)
	})

	switch cfg.Credentials.ClientCredentialsType {
	case config.AzureClientCredentialsTypeSecret:
		return p.getSecretCredentialsConfig(cfg)
	case config.AzureClientCredentialsTypeCertificate:
		return p.getCertificateCredentialsConfig(cfg)
	case config.AzureClientCredentialsTypeUsernamePassword:
		if cfg.Credentials.ClientUsername == "" || cfg.Credentials.ClientPassword == "" {
			return nil, ErrIncompleteUsernamePassword
		}
		return p.getUsernamePasswordCredentialsConfig(cfg)
	case "", config.AzureClientCredentialsTypeManagedIdentity, config.AzureClientCredentialsTypeARMTemplate, config.AzureClientCredentialsTypeManual:
		return p.getDefaultCredentialsConfig()
	}

	return nil, ErrWrongCredentialsType
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

func (p *ConfigProvider) getSecretCredentialsConfig(cfg config.AzureConfig) (*AzureFactoryConfig, error) {
	creds, err := p.AuthProvider.FindClientSecretCredentials(
		cfg.Credentials.TenantID,
		cfg.Credentials.ClientID,
		cfg.Credentials.ClientSecret,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret credentials: %w", err)
	}

	return &AzureFactoryConfig{
		Credentials: creds,
	}, nil
}

func (p *ConfigProvider) getCertificateCredentialsConfig(cfg config.AzureConfig) (*AzureFactoryConfig, error) {
	creds, err := p.AuthProvider.FindCertificateCredential(
		cfg.Credentials.TenantID,
		cfg.Credentials.ClientID,
		cfg.Credentials.ClientCertificatePath,
		cfg.Credentials.ClientCertificatePassword,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret credentials: %w", err)
	}

	return &AzureFactoryConfig{
		Credentials: creds,
	}, nil
}

func (p *ConfigProvider) getUsernamePasswordCredentialsConfig(cfg config.AzureConfig) (*AzureFactoryConfig, error) {
	creds, err := p.AuthProvider.FindUsernamePasswordCredentials(
		cfg.Credentials.TenantID,
		cfg.Credentials.ClientID,
		cfg.Credentials.ClientUsername,
		cfg.Credentials.ClientPassword,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get username and password credentials: %w", err)
	}

	return &AzureFactoryConfig{
		Credentials: creds,
	}, nil
}

var (
	ErrWrongCredentialsType       = errors.New("wrong credentials type")
	ErrIncompleteUsernamePassword = errors.New("incomplete username and password credentials")
)
