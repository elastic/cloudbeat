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

package gcplib

import (
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"google.golang.org/api/option"
	"os"
	"reflect"
	"testing"
)

const (
	saCredentialsJSON = `{ "client_id": "test" }`
	saFilePath        = "sa-credentials.json"
)

func TestGetGcpClientConfig(t *testing.T) {
	f := createServiceAccountFile(t)
	defer os.Remove(f.Name())

	tests := []struct {
		name    string
		cfg     *config.Config
		want    []option.ClientOption
		wantErr bool
	}{
		{
			name: "Should return a GcpClientConfig using SA credentials file path",
			cfg: &config.Config{
				CloudConfig: config.CloudConfig{
					GcpCfg: config.GcpClientOpt{
						CredentialsFilePath: saFilePath,
					},
				},
			},
			want:    []option.ClientOption{option.WithCredentialsFile(saFilePath)},
			wantErr: false,
		},
		{
			name: "Should return an error due to invalid SA credentials file path",
			cfg: &config.Config{
				CloudConfig: config.CloudConfig{
					GcpCfg: config.GcpClientOpt{
						CredentialsFilePath: "invalid path",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should return a GcpClientConfig using SA credentials json",
			cfg: &config.Config{
				CloudConfig: config.CloudConfig{
					GcpCfg: config.GcpClientOpt{
						CredentialsJSON: saCredentialsJSON,
					},
				},
			},
			want:    []option.ClientOption{option.WithCredentialsJSON([]byte(saCredentialsJSON))},
			wantErr: false,
		},
		{
			name: "Should return an error due to invalid SA json",
			cfg: &config.Config{
				CloudConfig: config.CloudConfig{
					GcpCfg: config.GcpClientOpt{
						CredentialsJSON: "invalid json",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Should return client options with both credentials_file_path and credentials_json",
			cfg: &config.Config{
				CloudConfig: config.CloudConfig{
					GcpCfg: config.GcpClientOpt{
						CredentialsFilePath: saFilePath,
						CredentialsJSON:     saCredentialsJSON,
					},
				},
			},
			want: []option.ClientOption{
				option.WithCredentialsFile(saFilePath),
				option.WithCredentialsJSON([]byte(saCredentialsJSON)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGcpClientConfig(tt.cfg, logp.NewLogger("gcp credentials test"))
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGcpClientConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGcpClientConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// Creates a test sa account file to be used in the tests
func createServiceAccountFile(t *testing.T) *os.File {
	f, err := os.Create(saFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(saCredentialsJSON); err != nil {
		t.Fatal(err)
	}
	return f
}
