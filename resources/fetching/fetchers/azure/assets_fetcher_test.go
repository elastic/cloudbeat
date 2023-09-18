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
