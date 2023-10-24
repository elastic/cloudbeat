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

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

func TestK8sDataProvider_EnrichEvent(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		want    map[string]interface{}
	}{
		{
			name: "should return empty map",
			options: []Option{
				WithLogger(testhelper.NewLogger(t)),
			},
			want: map[string]interface{}{},
		}, {
			name: "should return cluster version",
			options: []Option{
				WithLogger(testhelper.NewLogger(t)),
				WithClusterVersion("test_version"),
			},
			want: map[string]interface{}{
				clusterVersionField: "test_version",
			},
		}, {
			name: "should return cluster name",
			options: []Option{
				WithLogger(testhelper.NewLogger(t)),
				WithClusterName("test_cluster"),
			},
			want: map[string]interface{}{
				clusterNameField: "test_cluster",
			},
		}, {
			name: "should return cluster id",
			options: []Option{
				WithLogger(testhelper.NewLogger(t)),
				WithClusterID("test_id"),
			},
			want: map[string]interface{}{
				clusterIdField: "test_id",
			},
		}, {
			name: "should return all fields",
			options: []Option{
				WithLogger(testhelper.NewLogger(t)),
				WithClusterID("test_id"),
				WithClusterName("test_cluster"),
				WithClusterVersion("test_version"),
			},
			want: map[string]interface{}{
				clusterIdField:      "test_id",
				clusterNameField:    "test_cluster",
				clusterVersionField: "test_version",
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
