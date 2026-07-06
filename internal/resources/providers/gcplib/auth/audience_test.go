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

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectNumberFromAudience(t *testing.T) {
	tests := []struct {
		name     string
		audience string
		wantNum  string
		wantOK   bool
	}{
		{
			name:     "empty audience returns false",
			audience: "",
			wantNum:  "",
			wantOK:   false,
		},
		{
			name:     "Workload Identity Federation audience format",
			audience: "//iam.googleapis.com/projects/123456/locations/global/workloadIdentityPools/test-pool/providers/test-provider",
			wantNum:  "123456",
			wantOK:   true,
		},
		{
			name:     "audience with no projects segment",
			audience: "//iam.googleapis.com/locations/global/workloadIdentityPools/test-pool",
			wantNum:  "",
			wantOK:   false,
		},
		{
			name:     "projects segment with empty next segment",
			audience: "//iam.googleapis.com/projects//locations/global",
			wantNum:  "",
			wantOK:   false,
		},
		{
			name:     "project number with surrounding whitespace is trimmed",
			audience: "prefix/projects/  987654  /locations/global",
			wantNum:  "987654",
			wantOK:   true,
		},
		{
			name:     "returns first project number when multiple projects segments",
			audience: "projects/111/locations/global/projects/222/locations/eu",
			wantNum:  "111",
			wantOK:   true,
		},
		{
			name:     "projects at end without number",
			audience: "//iam.googleapis.com/some/path/projects",
			wantNum:  "",
			wantOK:   false,
		},
		{
			name:     "single segment projects",
			audience: "projects/42",
			wantNum:  "42",
			wantOK:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNum, gotOK := projectNumberFromAudience(tt.audience)
			assert.Equal(t, tt.wantNum, gotNum)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}
