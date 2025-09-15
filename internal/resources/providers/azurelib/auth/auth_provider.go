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
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/elastic/cloudbeat/internal/config"
)

type AzureAuthProvider struct{}

type AzureAuthProviderAPI interface {
	FindDefaultCredentials(options *azidentity.DefaultAzureCredentialOptions) (*azidentity.DefaultAzureCredential, error)
	FindClientSecretCredentials(tenantID string, clientID string, clientSecret string, options *azidentity.ClientSecretCredentialOptions) (*azidentity.ClientSecretCredential, error)
	FindCertificateCredential(tenantID string, clientID string, certPath string, password string, options *azidentity.ClientCertificateCredentialOptions) (*azidentity.ClientCertificateCredential, error)
	FindClientAssertionCredentials(tenantID string, clientID string, options *azidentity.ClientAssertionCredentialOptions) (*azidentity.ClientAssertionCredential, error)
}

// FindDefaultCredentials is a wrapper around azidentity.NewDefaultAzureCredential to make it easier to mock
func (a *AzureAuthProvider) FindDefaultCredentials(options *azidentity.DefaultAzureCredentialOptions) (*azidentity.DefaultAzureCredential, error) {
	return azidentity.NewDefaultAzureCredential(options)
}

// FindClientSecretCredentials is a wrapper around azidentity.NewClientSecretCredential to make it easier to mock
func (a *AzureAuthProvider) FindClientSecretCredentials(tenantID string, clientID string, clientSecret string, options *azidentity.ClientSecretCredentialOptions) (*azidentity.ClientSecretCredential, error) {
	return azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, options)
}

// FindCertificateCredential is a wrapper around azidentity.NewClientCertificateCredential and azidentity.ParseCertificates that loads certificates and a private key, in PEM or PKCS12 format.
func (a *AzureAuthProvider) FindCertificateCredential(tenantID string, clientID string, certPath string, password string, options *azidentity.ClientCertificateCredentialOptions) (*azidentity.ClientCertificateCredential, error) {
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("error trying to read certificate file %s: %s", certPath, err.Error())
	}

	// ParseCertificates requires nil password if the private key isn't encrypted. It can't decrypt keys in PEM format.
	var pwd []byte
	if len(password) > 0 {
		pwd = []byte(password)
	}
	certs, key, err := azidentity.ParseCertificates(certData, pwd)
	if err != nil {
		return nil, fmt.Errorf("error trying to load certificate from %s: %s", certPath, err.Error())
	}

	return azidentity.NewClientCertificateCredential(tenantID, clientID, certs, key, options)
}

// FindClientAssertionCredentials is a wrapper around azidentity.NewClientAssertionCredential that loads JWT from environment variable, similar to cloud connectors pattern
func (a *AzureAuthProvider) FindClientAssertionCredentials(tenantID string, clientID string, options *azidentity.ClientAssertionCredentialOptions) (*azidentity.ClientAssertionCredential, error) {
	jwtFilePath := os.Getenv(config.CloudConnectorsJWTPathEnvVar)
	if jwtFilePath == "" {
		return nil, fmt.Errorf("environment variable %s is required for client assertion credentials", config.CloudConnectorsJWTPathEnvVar)
	}

	getAssertion := func(_ context.Context) (string, error) {
		return readJWTFromFile(jwtFilePath)
	}

	return azidentity.NewClientAssertionCredential(tenantID, clientID, getAssertion, options)
}

func readJWTFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error trying to read JWT file %s: %s", filePath, err.Error())
	}

	jwt := strings.TrimSpace(string(data))
	if jwt == "" {
		return "", fmt.Errorf("JWT file %s is empty", filePath)
	}

	// Basic validation - JWT should have 3 parts separated by dots
	parts := strings.Count(jwt, ".")
	if parts != 2 {
		return "", fmt.Errorf("invalid JWT format in file %s: expected 3 parts separated by dots, got %d", filePath, parts+1)
	}

	return jwt, nil
}
