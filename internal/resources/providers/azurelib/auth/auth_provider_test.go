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
	"os"
	"path/filepath"
	"testing"

	"github.com/elastic/beats/v7/x-pack/libbeat/common/identityfederation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/config"
)

func TestAzureAuthProvider_FindClientAssertionCredentials(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func(t *testing.T) string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Should successfully create credential when env var and file exist",
			setupEnv: func(t *testing.T) string {
				tempDir := t.TempDir()
				jwtFile := filepath.Join(tempDir, "jwt.token")
				require.NoError(t, os.WriteFile(jwtFile, []byte("eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.signature"), 0644))
				t.Setenv(identityfederation.WIIIssuerURLEnvVar, "")
				t.Setenv(config.CloudConnectorsJWTPathEnvVar, jwtFile)
				return jwtFile
			},
			expectError: false,
		},
		{
			name: "Should fail when neither WII nor the JWT file env var is set",
			setupEnv: func(t *testing.T) string {
				t.Setenv(identityfederation.WIIIssuerURLEnvVar, "")
				t.Setenv(config.CloudConnectorsJWTPathEnvVar, "")
				return ""
			},
			expectError: true,
			errorMsg:    "no id token source configured",
		},
		{
			name: "wii: Should create a WII credential when the issuer URL, cert and key are set",
			setupEnv: func(t *testing.T) string {
				t.Setenv(identityfederation.WIIIssuerURLEnvVar, "https://workload-identity-issuer.example.com")
				t.Setenv(identityfederation.WIISSLCertFileEnvVar, "/certs/wii-client.crt")
				t.Setenv(identityfederation.WIISSLKeyFileEnvVar, "/certs/wii-client.key")
				t.Setenv(config.CloudConnectorsJWTPathEnvVar, "")
				return ""
			},
			expectError: false,
		},
		{
			name: "wii: cert and key files are required and there is no silent legacy fallback",
			setupEnv: func(t *testing.T) string {
				t.Setenv(identityfederation.WIIIssuerURLEnvVar, "https://workload-identity-issuer.example.com")
				t.Setenv(identityfederation.WIISSLCertFileEnvVar, "")
				t.Setenv(identityfederation.WIISSLKeyFileEnvVar, "")
				t.Setenv(config.CloudConnectorsJWTPathEnvVar, "/path/token")
				return ""
			},
			expectError: true,
			errorMsg:    "WORKLOAD_IDENTITY_SSL_CERT_FILE",
		},
		{
			name: "wii: Should fail when the WII issuer URL is not https",
			setupEnv: func(t *testing.T) string {
				t.Setenv(identityfederation.WIIIssuerURLEnvVar, "http://workload-identity-issuer.example.com")
				t.Setenv(identityfederation.WIISSLCertFileEnvVar, "/certs/wii-client.crt")
				t.Setenv(identityfederation.WIISSLKeyFileEnvVar, "/certs/wii-client.key")
				return ""
			},
			expectError: true,
			errorMsg:    "must use the https scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv(t)

			provider := &AzureAuthProvider{}
			cred, err := provider.FindClientAssertionCredentials("tenant-id", "client-id", nil)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, cred)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cred)
			}
		})
	}

	t.Run("wii: tenant_id and client_id are required", func(t *testing.T) {
		t.Setenv(identityfederation.WIIIssuerURLEnvVar, "https://workload-identity-issuer.example.com")
		t.Setenv(identityfederation.WIISSLCertFileEnvVar, "/certs/wii-client.crt")
		t.Setenv(identityfederation.WIISSLKeyFileEnvVar, "/certs/wii-client.key")

		provider := &AzureAuthProvider{}
		_, err := provider.FindClientAssertionCredentials("", "", nil)
		require.ErrorContains(t, err, "tenant_id and client_id are required for WII identity federation")
	})
}

func TestReadJWTFromFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func() string
		expectError bool
		errorMsg    string
		expectedJWT string
	}{
		{
			name: "Should successfully read valid JWT",
			setupFile: func() string {
				tempDir := t.TempDir()
				jwtFile := filepath.Join(tempDir, "jwt.token")
				jwt := "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.signature"
				require.NoError(t, os.WriteFile(jwtFile, []byte(jwt), 0644))
				return jwtFile
			},
			expectError: false,
			expectedJWT: "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.signature",
		},
		{
			name: "Should fail when file does not exist",
			setupFile: func() string {
				return "/path/that/does/not/exist/jwt.token"
			},
			expectError: true,
			errorMsg:    "reading JWT file",
		},
		{
			name: "Should fail when file is empty",
			setupFile: func() string {
				tempDir := t.TempDir()
				jwtFile := filepath.Join(tempDir, "empty.token")
				require.NoError(t, os.WriteFile(jwtFile, []byte(""), 0644))
				return jwtFile
			},
			expectError: true,
			errorMsg:    "is empty",
		},
		{
			name: "Should fail when JWT format is invalid",
			setupFile: func() string {
				tempDir := t.TempDir()
				jwtFile := filepath.Join(tempDir, "invalid.token")
				require.NoError(t, os.WriteFile(jwtFile, []byte("invalid.jwt"), 0644))
				return jwtFile
			},
			expectError: true,
			errorMsg:    "invalid JWT in",
		},
		{
			name: "Should trim whitespace from JWT",
			setupFile: func() string {
				tempDir := t.TempDir()
				jwtFile := filepath.Join(tempDir, "jwt_with_whitespace.token")
				jwt := "  eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.signature  \n"
				require.NoError(t, os.WriteFile(jwtFile, []byte(jwt), 0644))
				return jwtFile
			},
			expectError: false,
			expectedJWT: "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtFile := tt.setupFile()

			jwt, err := identityfederation.AzureReadJWT(jwtFile)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedJWT, jwt)
			}
		})
	}
}
