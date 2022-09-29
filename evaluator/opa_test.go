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
	"testing"
	"time"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest/observer"
)

type DummyResource struct {
}

func (d *DummyResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{}, nil
}
func (d *DummyResource) GetData() any {
	return d
}
func (d *DummyResource) GetElasticCommonData() any {
	return d
}

func TestOpaEvaluator_decode(t *testing.T) {
	type args struct {
		result interface{}
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
				assert.Error(t, err, "expected to have an error")
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOpaEvaluatorWithDecisionLogs(t *testing.T) {
	ctx := context.Background()
	err := logp.DevelopmentSetup(logp.ToObserverOutput())
	assert.NoError(t, err)

	log := logp.NewLogger("opa_evaluator_test")

	tests := []struct {
		enabled  bool
		evals    int
		expected int
	}{
		{true, 1, 1},
		{true, 3, 3},
		{false, 1, 0},
		{false, 3, 0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("TestEvaluationsDecisionLogs %+v", tt), func(t *testing.T) {
			cfg := config.Config{
				Stream: config.Stream{
					Evaluator: config.EvaluatorConfig{
						DecisionLogs: tt.enabled,
					},
				},
			}

			e, err := NewOpaEvaluator(ctx, log, cfg)
			assert.NoError(t, err)

			for i := 0; i < tt.evals; i++ {
				_, err = e.Eval(ctx, fetching.ResourceInfo{
					Resource:      &DummyResource{},
					CycleMetadata: fetching.CycleMetadata{},
				})
				assert.NoError(t, err)
			}

			logs := findDecisionLogs()
			logp.ObserverLogs().TakeAll()
			assert.Len(t, logs, tt.expected)
			if tt.expected > 0 {
				assert.Contains(t, logs[0].ContextMap(), "decision_id")
			}
		})
	}
}

func findDecisionLogs() []observer.LoggedEntry {
	return logp.ObserverLogs().FilterMessage("Decision Log").TakeAll()
}
