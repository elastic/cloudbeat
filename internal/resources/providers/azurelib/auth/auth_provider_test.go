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
				t.Setenv(config.CloudConnectorsJWTPathEnvVar, jwtFile)
				return jwtFile
			},
			expectError: false,
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
}
