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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/infra/clog"
)

const (
	saCredentialsJSON   = `{ "client_id": "test" }`
	saFilePath          = "sa-credentials.json"
	testProjectId       = "test-project"
	testParentProjectId = "projects/test-project"
	testOrgId           = "test-organization"
	testParentOrgId     = "organizations/test-organization"
)

func TestGetGcpClientConfig(t *testing.T) {
	f := createServiceAccountFile(t)
	defer func() {
		require.NoError(t, os.Remove(f.Name()))
	}()

	tests := []struct {
		name         string
		cfg          []config.GcpConfig
		authProvider GoogleAuthProviderAPI
		want         []*GcpFactoryConfig
		wantErr      bool
	}{
		{
			name: "Should return a GcpClientConfig using SA credentials file path",
			cfg: []config.GcpConfig{
				{
					AccountType: config.SingleAccount,
					ProjectId:   testProjectId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsFilePath: saFilePath,
					},
				},
				{
					AccountType:    config.OrganizationAccount,
					OrganizationId: testOrgId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsFilePath: saFilePath,
					},
				},
			},
			authProvider: mockGoogleAuthProvider(nil),
			want: []*GcpFactoryConfig{
				{
					Parent:     testParentProjectId,
					ClientOpts: []option.ClientOption{option.WithCredentialsFile(saFilePath)},
				},
				{
					Parent:     testParentOrgId,
					ClientOpts: []option.ClientOption{option.WithCredentialsFile(saFilePath)},
				},
			},
			wantErr: false,
		},
		{
			name: "Should return an error due to invalid SA credentials file path",
			cfg: []config.GcpConfig{
				{
					AccountType: config.SingleAccount,
					ProjectId:   testProjectId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsFilePath: "invalid path",
					},
				},
				{
					AccountType:    config.OrganizationAccount,
					OrganizationId: testOrgId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsFilePath: "invalid path",
					},
				},
			},
			authProvider: mockGoogleAuthProvider(nil),
			want:         nil,
			wantErr:      true,
		},
		{
			name: "Should return a GcpClientConfig using SA credentials json",
			cfg: []config.GcpConfig{
				{
					AccountType: config.SingleAccount,
					ProjectId:   testProjectId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsJSON: saCredentialsJSON,
					},
				},
				{
					AccountType:    config.OrganizationAccount,
					OrganizationId: testOrgId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsJSON: saCredentialsJSON,
					},
				},
			},
			authProvider: mockGoogleAuthProvider(nil),
			want: []*GcpFactoryConfig{
				{
					Parent:     testParentProjectId,
					ClientOpts: []option.ClientOption{option.WithCredentialsJSON([]byte(saCredentialsJSON))},
				},
				{
					Parent:     testParentOrgId,
					ClientOpts: []option.ClientOption{option.WithCredentialsJSON([]byte(saCredentialsJSON))},
				},
			},
			wantErr: false,
		},
		{
			name: "Should return an error due to invalid SA json",
			cfg: []config.GcpConfig{
				{
					AccountType: config.SingleAccount,
					ProjectId:   testProjectId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsJSON: "invalid json",
					},
				},
				{
					AccountType:    config.OrganizationAccount,
					OrganizationId: testOrgId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsJSON: "invalid json",
					},
				},
			},
			authProvider: mockGoogleAuthProvider(nil),
			want:         nil,
			wantErr:      true,
		},
		{
			name: "Should return client options with both credentials_file_path and credentials_json",
			cfg: []config.GcpConfig{
				{
					AccountType: config.SingleAccount,
					ProjectId:   testProjectId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsFilePath: saFilePath,
						CredentialsJSON:     saCredentialsJSON,
					},
				},
				{
					AccountType:    config.OrganizationAccount,
					OrganizationId: testOrgId,
					GcpClientOpt: config.GcpClientOpt{
						CredentialsFilePath: saFilePath,
						CredentialsJSON:     saCredentialsJSON,
					},
				},
			},
			authProvider: mockGoogleAuthProvider(nil),
			want: []*GcpFactoryConfig{
				{
					Parent: testParentProjectId,
					ClientOpts: []option.ClientOption{
						option.WithCredentialsFile(saFilePath),
						option.WithCredentialsJSON([]byte(saCredentialsJSON)),
					},
				},
				{
					Parent: testParentOrgId,
					ClientOpts: []option.ClientOption{
						option.WithCredentialsFile(saFilePath),
						option.WithCredentialsJSON([]byte(saCredentialsJSON)),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Should return nil and use application default credentials",
			cfg: []config.GcpConfig{
				{
					AccountType:  config.SingleAccount,
					ProjectId:    testProjectId,
					GcpClientOpt: config.GcpClientOpt{},
				},
				{
					AccountType:    config.OrganizationAccount,
					OrganizationId: testOrgId,
					GcpClientOpt:   config.GcpClientOpt{},
				},
			},
			authProvider: mockGoogleAuthProvider(nil),
			want: []*GcpFactoryConfig{
				{
					Parent:     testParentProjectId,
					ClientOpts: nil,
				},
				{
					Parent:     testParentOrgId,
					ClientOpts: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "Should return the project id retrieved from configuration",
			cfg: []config.GcpConfig{
				{
					AccountType:  config.SingleAccount,
					ProjectId:    testProjectId,
					GcpClientOpt: config.GcpClientOpt{},
				},
				{
					AccountType:    config.OrganizationAccount,
					OrganizationId: testOrgId,
					GcpClientOpt:   config.GcpClientOpt{},
				},
			},
			authProvider: mockGoogleAuthProvider(nil),
			want: []*GcpFactoryConfig{
				{
					Parent:     testParentProjectId,
					ClientOpts: nil,
				},
				{
					Parent:     testParentOrgId,
					ClientOpts: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "Should return nil due to error getting application default credentials",
			cfg: []config.GcpConfig{
				{
					AccountType:  config.SingleAccount,
					GcpClientOpt: config.GcpClientOpt{},
				},
			},
			authProvider: mockGoogleAuthProvider(errors.New("fail to retrieve ADC")),
			want:         nil,
			wantErr:      true,
		},
		{
			name: "returns an error due to missing org ID in org account type",
			cfg: []config.GcpConfig{
				{
					AccountType:  config.OrganizationAccount,
					GcpClientOpt: config.GcpClientOpt{},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "returns an error due to missing/unknown account type",
			cfg:     []config.GcpConfig{{GcpClientOpt: config.GcpClientOpt{}}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		p := ConfigProvider{
			AuthProvider: tt.authProvider,
		}
		t.Run(tt.name, func(t *testing.T) {
			for idx, cfg := range tt.cfg {
				got, err := p.GetGcpClientConfig(t.Context(), cfg, clog.NewLogger("gcp credentials test"))
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}

				if tt.want != nil {
					assert.Equal(t, tt.want[idx], got)
				}
				if tt.want == nil {
					assert.Nil(t, got)
				}
			}
		})
	}
}

// Creates a test sa account file to be used in the tests
func createServiceAccountFile(t *testing.T) *os.File {
	f, err := os.Create(saFilePath)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()

	_, err = f.WriteString(saCredentialsJSON)
	require.NoError(t, err)
	return f
}

func mockGoogleAuthProvider(err error) *MockGoogleAuthProviderAPI {
	googleProviderAPI := &MockGoogleAuthProviderAPI{}
	on := googleProviderAPI.EXPECT().FindDefaultCredentials(mock.Anything)
	if err == nil {
		on.Return(
			&google.Credentials{ProjectID: testProjectId},
			nil,
		)
	} else {
		on.Return(nil, err)
	}
	return googleProviderAPI
}
