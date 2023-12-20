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

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

type DummyResource struct {
}

func (d *DummyResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{}, nil
}
func (d *DummyResource) GetData() any {
	return d
}
func (d *DummyResource) GetElasticCommonData() (map[string]any, error) {
	return nil, nil
}

type OpaTestSuite struct {
	suite.Suite
	log *logp.Logger
}

func TestOpaTestSuite(t *testing.T) {
	testhelper.SkipLong(t)

	s := new(OpaTestSuite)
	s.log = testhelper.NewLogger(t)

	suite.Run(t, s)
}

func (s *OpaTestSuite) SetupSuite() {
	err := logp.TestingSetup(logp.ToObserverOutput())
	s.Require().NoError(err)
}

func (s *OpaTestSuite) TestOpaEvaluator_decode() {
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
		s.Run(tt.name, func() {
			o := &OpaEvaluator{}
			if tt.args.now != nil {
				now = tt.args.now
			}
			got, err := o.decode(tt.args.result)
			if tt.wantErr {
				s.Require().Error(err, "expected to have an error")
				return
			}
			s.Require().NoError(err)
			s.Equal(tt.want, got)
		})
	}
}

func (s *OpaTestSuite) TestOpaEvaluatorWithDecisionLogs() {
	testhelper.SkipLong(s.T())

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
			s.Require().NoError(err)

			for i := 0; i < tt.evals; i++ {
				_, err = e.Eval(ctx, fetching.ResourceInfo{
					Resource:      &DummyResource{},
					CycleMetadata: cycle.Metadata{},
				})
				s.Require().NoError(err)
			}

			logs := findDecisionLogs()
			logp.ObserverLogs().TakeAll()
			s.Len(logs, tt.expected)
			if tt.expected > 0 {
				s.Contains(logs[0].ContextMap(), "decision_id")
				s.Equal(zapcore.DebugLevel, logs[0].Level)
			}
		})
	}
}

func (s *OpaTestSuite) getTestConfig() *config.Config {
	path, err := filepath.Abs("bundle.tar.gz")
	s.Require().NoError(err)
	_, err = os.Stat(path)
	s.Require().NoError(err)
	return &config.Config{
		BundlePath: path,
	}
}

func findDecisionLogs() []observer.LoggedEntry {
	return logp.ObserverLogs().FilterMessageSnippet("Decision Log").TakeAll()
}
