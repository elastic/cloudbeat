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

package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
)

var (
	log = logp.NewLogger("cloudbeat_config_test_suite")
)

func TestStep(t *testing.T) {
	tests := []struct {
		name    string
		fn      func(context.Context, int) (float64, error)
		input   int
		wantLen int
	}{
		{
			name:    "Should receive value from output channel",
			fn:      func(context context.Context, i int) (float64, error) { return float64(i), nil },
			input:   1,
			wantLen: 1,
		},
		{
			name:    "Pipeline function returns error - no value received",
			fn:      func(context context.Context, i int) (float64, error) { return 0, errors.New("some error") },
			input:   2,
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputChannel := make(chan int)
			outCh := Step(context.Background(), log, inputChannel, tt.fn)
			inputChannel <- tt.input
			close(inputChannel)

			results := testhelper.CollectResourcesBlocking(outCh)
			assert.Equal(t, tt.wantLen, len(results))
		})
	}
}
