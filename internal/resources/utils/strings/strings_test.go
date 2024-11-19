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

package strings

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "nil",
			args: nil,
			want: "",
		},
		{
			name: "empty",
			args: []string{},
			want: "",
		},
		{
			name: "first element",
			args: []string{"one", "two"},
			want: "one",
		},
		{
			name: "other element",
			args: []string{"", "", "", "some element"},
			want: "some element",
		},
		{
			name: "space",
			args: []string{"", " ", "xxx"},
			want: " ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FirstNonEmpty(tt.args...); got != tt.want {
				t.Errorf("FirstNonEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromMap(t *testing.T) {
	tests := map[string]struct {
		input    map[string]any
		inputKey string
		expected string
	}{
		"nil map": {
			input:    nil,
			inputKey: "test",
			expected: "",
		},
		"non existing key": {
			input:    map[string]any{"test": "value"},
			inputKey: "foo",
			expected: "",
		},
		"existing key with nil value": {
			input:    map[string]any{"test": nil},
			inputKey: "test",
			expected: "",
		},
		"non string value": {
			input:    map[string]any{"test": 1},
			inputKey: "test",
			expected: "",
		},
		"string value": {
			input:    map[string]any{"test": "value"},
			inputKey: "test",
			expected: "value",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := FromMap(tc.input, tc.inputKey)
			require.Equal(t, tc.expected, got)
		})
	}
}
