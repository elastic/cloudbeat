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

package dataprovider

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrichSuccess(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		expected map[string]string
	}{
		{
			name: "single value",
			data: map[string]interface{}{
				"a-field": "a-value",
			},
			expected: map[string]string{
				"a-field": "a-value",
			},
		},
		{
			name: "multiple values",
			data: map[string]interface{}{
				"some-field":  "some-value",
				"other-field": "other-value",
			},
			expected: map[string]string{
				"some-field":  "some-value",
				"other-field": "other-value",
			},
		},
		{
			name: "internal object",
			data: map[string]interface{}{
				"some-field": "some-value",
				"more-fields": map[string]interface{}{
					"internal-field": "internal-value",
				},
			},
			expected: map[string]string{
				"some-field":                 "some-value",
				"more-fields.internal-field": "internal-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := NewMockElasticCommonDataProvider(t)
			dp.EXPECT().GetElasticCommonData().Return(tt.data, nil)

			ev := &beat.Event{
				Fields: map[string]interface{}{},
			}
			err := NewEnricher(dp).EnrichEvent(ev)
			require.NoError(t, err)

			for key, expectedValue := range tt.expected {
				actualValue, err := ev.GetValue(key)
				require.NoError(t, err)
				assert.Equal(t, expectedValue, actualValue)
			}
		})
	}
}

func TestEnrichError(t *testing.T) {
	dp := NewMockElasticCommonDataProvider(t)
	dp.EXPECT().GetElasticCommonData().Return(nil, assert.AnError)

	ev := &beat.Event{
		Fields: map[string]interface{}{},
	}
	err := NewEnricher(dp).EnrichEvent(ev)
	require.Error(t, err)
}
