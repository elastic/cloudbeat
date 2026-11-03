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

package awslib

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestWithGovCloudFallback(t *testing.T) {
	errFirst := errors.New("first error")
	errRetry := errors.New("retry error")

	tests := []struct {
		name          string
		defaultRegion string
		results       []error
		wantCalls     int
		wantRegion    string
		wantErrs      []error
	}{
		{
			name:       "first call succeeds",
			results:    []error{nil},
			wantCalls:  1,
			wantRegion: "",
		},
		{
			name:       "retry in govcloud succeeds",
			results:    []error{errFirst, nil},
			wantCalls:  2,
			wantRegion: DefaultGovRegion,
		},
		{
			name:       "retry in govcloud fails and keeps both errors",
			results:    []error{errFirst, errRetry},
			wantCalls:  2,
			wantRegion: DefaultGovRegion,
			wantErrs:   []error{errFirst, errRetry},
		},
		{
			name:          "no retry when the region is set",
			defaultRegion: "eu-west-1",
			results:       []error{errFirst},
			wantCalls:     1,
			wantRegion:    "eu-west-1",
			wantErrs:      []error{errFirst},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region := tt.defaultRegion
			calls := 0
			err := WithGovCloudFallback(testhelper.NewLogger(t), &region, func() error {
				calls++
				return tt.results[calls-1]
			})

			assert.Equal(t, tt.wantCalls, calls)
			assert.Equal(t, tt.wantRegion, region)
			if len(tt.wantErrs) == 0 {
				require.NoError(t, err)
			}
			for _, wantErr := range tt.wantErrs {
				require.ErrorIs(t, err, wantErr)
			}
		})
	}
}
