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

package fetchers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getString(t *testing.T) {
	tests := []struct {
		name string
		data map[string]any
		key  string
		want string
	}{
		{
			name: "nil map",
			data: nil,
			key:  "key",
			want: "",
		},
		{
			name: "key does not exist",
			data: map[string]any{"key": "value"},
			key:  "other-key",
			want: "",
		},
		{
			name: "wrong type",
			data: map[string]any{"key": 1},
			key:  "key",
			want: "",
		},
		{
			name: "correct value",
			data: map[string]any{"key": "value", "other-key": 1},
			key:  "key",
			want: "value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getString(tt.data, tt.key), "getString(%v, %s) = %s", tt.data, tt.key, tt.want)
		})
	}
}
