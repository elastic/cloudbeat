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

package common

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/version"
)

func Test_CommonDataProvider_GetElasticCommonData(t *testing.T) {
	tests := []struct {
		name string
		info version.CloudbeatVersionInfo
		cfg  *config.Config
		want map[string]any
	}{
		{
			name: "should return empty map",
			info: version.CloudbeatVersionInfo{},
			cfg:  nil,
			want: map[string]any{},
		}, {
			name: "should return cloudbeat version",
			info: version.CloudbeatVersionInfo{
				Version: version.Version{Version: "test_version"},
			},
			want: map[string]any{
				"cloudbeat.version": "test_version",
			},
		}, {
			name: "should return policy version",
			info: version.CloudbeatVersionInfo{
				Policy: version.Version{Version: "test_version"},
			},
			want: map[string]any{
				"cloudbeat.policy.version": "test_version",
			},
		}, {
			name: "should return full version",
			info: version.CloudbeatVersionInfo{
				Version: version.Version{Version: "test_cloudbeat_version"},
				Policy:  version.Version{Version: "test_policy_version"},
			},
			want: map[string]any{
				"cloudbeat.policy.version": "test_policy_version",
				"cloudbeat.version":        "test_cloudbeat_version",
			},
		}, {
			name: "should return package policy id and revision",
			info: version.CloudbeatVersionInfo{},
			cfg: &config.Config{
				PackagePolicyId:       "test_package_policy_id",
				PackagePolicyRevision: 1,
			},
			want: map[string]any{
				"cloud_security_posture.package_policy.id":       "test_package_policy_id",
				"cloud_security_posture.package_policy.revision": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.info, tt.cfg)
			require.NoError(t, err)

			ev := &beat.Event{
				Fields: map[string]any{},
			}

			err = dataprovider.NewEnricher(p).EnrichEvent(ev)
			require.NoError(t, err)

			fl := ev.Fields.Flatten()

			assert.Len(t, fl, len(tt.want))
			for key, expectedValue := range tt.want {
				actualValue := fl[key]
				assert.Equal(t, expectedValue, actualValue)
			}
		})
	}
}
