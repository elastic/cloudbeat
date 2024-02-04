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

package k8s

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
)

func TestK8sDataProvider_EnrichEvent(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		want    map[string]any
	}{
		{
			name:    "should return empty map",
			options: []Option{},
			want: map[string]any{
				orchestratorType: "kubernetes",
			},
		}, {
			name: "should return cluster version",
			options: []Option{
				WithClusterVersion("test_version"),
			},
			want: map[string]any{
				clusterVersionField: "test_version",
				orchestratorType:    "kubernetes",
			},
		}, {
			name: "should return cluster name",
			options: []Option{
				WithClusterName("test_cluster"),
			},
			want: map[string]any{
				clusterNameField: "test_cluster",
				orchestratorType: "kubernetes",
			},
		}, {
			name: "should return cluster id",
			options: []Option{
				WithClusterID("test_id"),
			},
			want: map[string]any{
				clusterIdField:   "test_id",
				orchestratorType: "kubernetes",
			},
		}, {
			name: "should return all fields",
			options: []Option{
				WithClusterID("test_id"),
				WithClusterName("test_cluster"),
				WithClusterVersion("test_version"),
			},
			want: map[string]any{
				clusterIdField:      "test_id",
				clusterNameField:    "test_cluster",
				clusterVersionField: "test_version",
				orchestratorType:    "kubernetes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := New(tt.options...)
			e := &beat.Event{
				Fields: mapstr.M{},
			}
			err := k.EnrichEvent(e, fetching.ResourceMetadata{})
			require.NoError(t, err)

			fl := e.Fields.Flatten()

			assert.Len(t, fl, len(tt.want))
			for key, expectedValue := range tt.want {
				actualValue, err := fl.GetValue(key)
				require.NoError(t, err)
				assert.Equal(t, expectedValue, actualValue)
			}
		})
	}
}
