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
	"path/filepath"
	"testing"
	"time"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
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

type OpaTestSuite struct {
	suite.Suite
	log *logp.Logger
}

func TestOpaTestSuite(t *testing.T) {
	s := new(OpaTestSuite)
	s.log = logp.NewLogger("opa_evaluator_test")

	suite.Run(t, s)
}

func (s *OpaTestSuite) SetupSuite() {
	err := logp.TestingSetup(logp.ToObserverOutput())
	s.NoError(err)
}

func (s *OpaTestSuite) TestOpaEvaluator_decode() {
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
		s.Run(tt.name, func() {
			o := &OpaEvaluator{}
			if tt.args.now != nil {
				now = tt.args.now
			}
			got, err := o.decode(tt.args.result)
			if tt.wantErr {
				s.Error(err, "expected to have an error")
				return
			}
			s.NoError(err)
			s.Equal(tt.want, got)
		})
	}
}

func (s *OpaTestSuite) TestOpaEvaluatorWithDecisionLogs() {
	ctx := context.Background()
	tests := []struct {
		evals    int
		expected int
	}{
		{1, 1},
		{3, 3},
	}

	for _, tt := range tests {
		s.Run(fmt.Sprintf("TestEvaluationsDecisionLogs %+v", tt), func() {
			cfg := s.getTestConfig()
			e, err := NewOpaEvaluator(ctx, s.log, cfg)
			s.NoError(err)

			for i := 0; i < tt.evals; i++ {
				_, err = e.Eval(ctx, fetching.ResourceInfo{
					Resource:      &DummyResource{},
					CycleMetadata: fetching.CycleMetadata{},
				})
				s.NoError(err)
			}

			logs := findDecisionLogs()
			logp.ObserverLogs().TakeAll()
			s.Len(logs, tt.expected)
			if tt.expected > 0 {
				s.Contains(logs[0].ContextMap(), "decision_id")
				s.Equal(logs[0].Level, zapcore.DebugLevel)
			}
		})
	}
}

func (s *OpaTestSuite) getTestConfig() config.Config {
	path, err := filepath.Abs("bundle.tar.gz")
	s.NoError(err)
	return config.Config{
		Stream: config.Stream{
			BundlePath: path,
		},
	}
}

func findDecisionLogs() []observer.LoggedEntry {
	return logp.ObserverLogs().FilterMessage("Decision Log").TakeAll()
}
