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
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/identityfederation"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/infra/clog"
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

// FindClientAssertionCredentials creates an Azure credential using a JWT identity token as the client assertion.
// It reads the env vars set by the agentless controller. Two paths are supported:
//
// WII path (WORKLOAD_IDENTITY_ISSUER_URL is set):
//   - JWT from the workload-identity-issuer, presented directly to the customer's federated identity credential.
//
// OIDC path (CLOUD_CONNECTORS_ID_TOKEN_FILE is set):
//   - JWT read from the token file, re-read on each token refresh via the shared beats azure package.
func (a *AzureAuthProvider) FindClientAssertionCredentials(tenantID string, clientID string, options *azidentity.ClientAssertionCredentialOptions) (*azidentity.ClientAssertionCredential, error) {
	logger := clog.NewLogger("FindClientAssertionCredentials")

	wiiIssuerURL := os.Getenv(identityfederation.WIIIssuerURLEnvVar)
	wiiCertFile := os.Getenv(identityfederation.WIISSLCertFileEnvVar)
	wiiKeyFile := os.Getenv(identityfederation.WIISSLKeyFileEnvVar)

	if wiiIssuerURL != "" {
		logger.Info("Switching credentials provider to Identity Federation (WII)")
		if wiiCertFile == "" || wiiKeyFile == "" {
			return nil, errors.New("WORKLOAD_IDENTITY_SSL_CERT_FILE and WORKLOAD_IDENTITY_SSL_KEY_FILE must be set when WORKLOAD_IDENTITY_ISSUER_URL is configured")
		}
		if tenantID == "" || clientID == "" {
			return nil, errors.New("tenant_id and client_id are required for WII identity federation")
		}
		tokenSource, err := identityfederation.NewWIITokenSource(wiiIssuerURL, wiiCertFile, wiiKeyFile, "api://AzureADTokenExchange", logger.Logger)
		if err != nil {
			return nil, fmt.Errorf("configuring WII token source: %w", err)
		}
		return identityfederation.AzureNewWIIClientAssertionCredential(identityfederation.AzureWIIParams{
			TenantID:    tenantID,
			ClientID:    clientID,
			TokenSource: tokenSource,
			Options:     options,
		})
	}

	jwtFilePath := os.Getenv(config.CloudConnectorsJWTPathEnvVar)
	if jwtFilePath == "" {
		return nil, errors.New("no id token source configured (set WORKLOAD_IDENTITY_ISSUER_URL or CLOUD_CONNECTORS_ID_TOKEN_FILE)")
	}

	logger.Info("Switching credentials provider to Identity Federation (OIDC)")
	return identityfederation.AzureNewClientAssertionCredential(identityfederation.AzureParams{
		TenantID:    tenantID,
		ClientID:    clientID,
		JWTFilePath: jwtFilePath,
		Options:     options,
	})
}
