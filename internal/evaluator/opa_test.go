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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type DummyResource struct{}

func (d *DummyResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{}, nil
}

func (d *DummyResource) GetData() any {
	return d
}

func (d *DummyResource) GetIds() []string {
	return nil
}

func (d *DummyResource) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}

func TestOpaEvaluator_decode(t *testing.T) {
	type args struct {
		result any
		now    func() time.Time
	}
	tests := []struct {
		name    string
		args    args
		want    RuleResult
		wantErr bool
	}{
		{
			name: "Should have sequence number",
			args: args{
				now: func() time.Time {
					return time.Unix(1, 0)
				},
			},
			want: RuleResult{
				Metadata: Metadata{
					CreatedAt: time.Unix(1, 0),
				},
			},
			wantErr: false,
		},
	}
	n := now
	for _, tt := range tests {
		now = n
		t.Run(tt.name, func(t *testing.T) {
			o := &OpaEvaluator{}
			if tt.args.now != nil {
				now = tt.args.now
			}
			got, err := o.decode(tt.args.result)
			if tt.wantErr {
				require.Error(t, err, "expected to have an error")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOpaEvaluatorWithDecisionLogs(t *testing.T) {
	testhelper.SkipLong(t)

	ctx := context.Background()
	tests := []struct {
		evals    int
		expected int
	}{
		{1, 1},
		{3, 3},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("TestEvaluationsDecisionLogs %+v", tt), func(t *testing.T) {
			testCfg := getTestConfig(t)
			log, observer := testhelper.NewObserverLogger(t)

			e, err := NewOpaEvaluator(ctx, log, testCfg)
			require.NoError(t, err)

			for i := 0; i < tt.evals; i++ {
				_, err = e.Eval(ctx, fetching.ResourceInfo{
					Resource:      &DummyResource{},
					CycleMetadata: cycle.Metadata{},
				})
				require.NoError(t, err)
			}

			logs := observer.FilterMessageSnippet("Decision Log").TakeAll()
			require.Len(t, logs, tt.expected)
			if tt.expected > 0 {
				assert.Contains(t, logs[0].ContextMap(), "decision_id")
				assert.Equal(t, zapcore.DebugLevel, logs[0].Level)
			}
		})
	}
}

func getTestConfig(t *testing.T) *config.Config {
	t.Helper()
	path, err := filepath.Abs("../../bundle.tar.gz")
	require.NoError(t, err)
	_, err = os.Stat(path)
	require.NoError(t, err)
	return &config.Config{
		BundlePath: path,
	}
}
