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

package evaluator

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKubeServiceAccountsEvidence checks that the CIS K8s 5.1.6 finding for a Pod and for a
// ServiceAccount carries the evaluated values in its evidence.
func TestKubeServiceAccountsEvidence(t *testing.T) {
	bundleDir, err := filepath.Abs("../../security-policies/bundle")
	require.NoError(t, err)

	query, err := rego.New(
		rego.Query("data.compliance.cis_k8s.rules.cis_5_1_6.finding"),
		rego.Load([]string{bundleDir}, nil),
	).PrepareForEval(context.Background())
	require.NoError(t, err)

	podInput := map[string]any{
		"type": "k8s_object",
		"resource": map[string]any{
			"kind":     "Pod",
			"metadata": map[string]any{"name": "pod"},
			"spec": map[string]any{
				"serviceAccount":               "app-sa",
				"serviceAccountName":           "app-sa",
				"automountServiceAccountToken": true,
			},
		},
	}
	saInput := map[string]any{
		"type": "k8s_object",
		"resource": map[string]any{
			"kind":                         "ServiceAccount",
			"metadata":                     map[string]any{"name": "default"},
			"automountServiceAccountToken": true,
		},
	}

	tests := []struct {
		name     string
		input    map[string]any
		expected []any
		key      string
	}{
		{
			name:  "pod evidence",
			input: podInput,
			key:   "serviceAccounts",
			expected: []any{map[string]any{
				"serviceAccountName":           "app-sa",
				"automountServiceAccountToken": true,
			}},
		},
		{
			name:  "service account evidence",
			input: saInput,
			key:   "serviceAccount",
			expected: []any{map[string]any{
				"name":                         "default",
				"automountServiceAccountToken": true,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs, err := query.Eval(context.Background(), rego.EvalInput(tt.input))
			require.NoError(t, err)
			require.Len(t, rs, 1)
			require.Len(t, rs[0].Expressions, 1)

			finding, ok := rs[0].Expressions[0].Value.(map[string]any)
			require.True(t, ok)
			evidence, ok := finding["evidence"].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, tt.expected, evidence[tt.key])
		})
	}
}
