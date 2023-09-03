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
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/version"
	"github.com/stretchr/testify/assert"
)

func Test_CommonDataProvider_GetElasticCommonData(t *testing.T) {
	tests := []struct {
		name string
		info version.CloudbeatVersionInfo
		want map[string]interface{}
	}{
		{
			name: "should return empty map",
			info: version.CloudbeatVersionInfo{},
			want: map[string]interface{}{},
		}, {
			name: "should return cloudbeat version",
			info: version.CloudbeatVersionInfo{
				Version: version.CloudbeatVersion(),
			},
			want: map[string]interface{}{
				"cloudbeat.version": map[string]interface{}{},
			},
		}, {
			name: "should return policy version",
			info: version.CloudbeatVersionInfo{
				Policy: version.PolicyVersion(),
			},
			want: map[string]interface{}{
				"cloudbeat.policy.version": "v1.5.4",
			},
		}, {
			name: "should return kubernetes version",
			info: version.CloudbeatVersionInfo{
				Policy:  version.PolicyVersion(),
				Version: version.CloudbeatVersion(),
			},
			want: map[string]interface{}{
				"cloudbeat": map[string]interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(testhelper.NewLogger(t), tt.info)
			ev := &beat.Event{
				Fields: map[string]interface{}{},
			}

			err := dataprovider.NewEnricher(p).EnrichEvent(ev)
			assert.NoError(t, err)

			for key, expectedValue := range tt.want {
				actualValue, err := ev.GetValue(key)
				assert.NoError(t, err)
				assert.Equal(t, expectedValue, actualValue)
			}
		})
	}
}
