package bundle

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestPolicyMap(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		fsys         fs.FS
		filePrefixes []string
		expected     []string
	}{
		{
			fsys: fstest.MapFS{
				"cis/main.rego":      {},
				"cis/main.test.rego": {},
				"cis/file.rego":      {},
				"cis/data.json":      {},
				"cis/data.yaml":      {},
			},
			filePrefixes: []string{
				"cis",
			},
			expected: []string{
				"cis/main.rego",
				"cis/file.rego",
				"cis/data.json",
				"cis/data.yaml",
			},
		},
	}

	for _, test := range tests {
		result, err := createPolicyMap(test.fsys, test.filePrefixes)
		assert.NoError(err)
		assert.Equal(len(test.expected), len(result))

		for _, file := range test.expected {
			_, ok := result[file]
			assert.True(ok)
		}
	}
}

func TestFileInclude(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		input        string
		filePrefixes []string
		expected     bool
	}{
		{
			input:        "cis/main.test.rego",
			filePrefixes: []string{},
			expected:     false,
		},
		{
			input:        "main.rego",
			filePrefixes: []string{"cis/"},
			expected:     false,
		},
		{
			input:        "cis/main.test.rego",
			filePrefixes: []string{"cis/"},
			expected:     false,
		},
		{
			input:        "cis/main.rego",
			filePrefixes: []string{"cis/"},
			expected:     true,
		},
		{
			input:        "cis/data.json",
			filePrefixes: []string{"cis/"},
			expected:     true,
		},
		{
			input:        "cis/data.yaml",
			filePrefixes: []string{"cis/"},
			expected:     true,
		},
		{
			input:        "cis/data.yml",
			filePrefixes: []string{"cis/"},
			expected:     true,
		},
		{
			input:        "cis/main.rego",
			filePrefixes: []string{"k8s/", "cis/"},
			expected:     true,
		},
	}

	for _, test := range tests {
		result := includeFile(test.filePrefixes, test.input)
		assert.Equal(test.expected, result)
	}
}
