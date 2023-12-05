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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestDereference(t *testing.T) {
	tests := []struct {
		name string
		s    *string
		want string
	}{
		{
			name: "nil",
			s:    nil,
			want: "",
		},
		{
			name: "empty",
			s:    aws.String(""),
			want: "",
		},
		{
			name: "something",
			s:    aws.String("something"),
			want: "something",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Dereference(tt.s); got != tt.want {
				t.Errorf("Dereference() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
	tests := []struct {
		name string
		data map[string]any
		key  string
		want string
	}{
		{
			name: "everything empty",
		},
		{
			name: "nil",
			data: nil,
			key:  "key",
			want: "",
		},
		{
			name: "empty map",
			data: make(map[string]any),
			key:  "key",
			want: "",
		},
		{
			name: "wrong key",
			data: map[string]any{"a": "b"},
			key:  "key",
			want: "",
		},
		{
			name: "nil value",
			data: map[string]any{"key": nil},
			key:  "key",
			want: "",
		},
		{
			name: "wrong type",
			data: map[string]any{"key": map[any]any{"other": "map"}},
			key:  "key",
			want: "",
		},
		{
			name: "success",
			data: map[string]any{"key": "value"},
			key:  "key",
			want: "value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FromMap(tt.data, tt.key))
		})
	}
}
