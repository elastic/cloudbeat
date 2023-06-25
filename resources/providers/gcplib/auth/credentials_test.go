package gcplib

import (
	"github.com/elastic/cloudbeat/config"
	"google.golang.org/api/option"
	"reflect"
	"testing"
)

func TestGetGcpClientConfig(t *testing.T) {
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
						CredentialsFilePath: "sa-credentials.json",
					},
				},
			},
			want:    []option.ClientOption{option.WithCredentialsFile("sa-credentials.json")},
			wantErr: false,
		},
		{
			name: "Should return a GcpClientConfig using SA credentials json",
			cfg: &config.Config{
				CloudConfig: config.CloudConfig{
					GcpCfg: config.GcpClientOpt{
						CredentialsJSON: "test-json-content",
					},
				},
			},
			want:    []option.ClientOption{option.WithCredentialsJSON([]byte("test-json-content"))},
			wantErr: false,
		},
		{
			name: "Should return error when both credentials_file_path and credentials_json specified",
			cfg: &config.Config{
				CloudConfig: config.CloudConfig{
					GcpCfg: config.GcpClientOpt{
						CredentialsFilePath: "sa-credentials.json",
						CredentialsJSON:     "test-json-content",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGcpClientConfig(tt.cfg)
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
