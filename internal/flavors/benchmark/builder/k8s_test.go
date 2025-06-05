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

package builder

import (
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/elastic/cloudbeat/internal/evaluator"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/uniqueness"
)

func TestK8sRun_ReturnEvents(t *testing.T) {
	testhelper.SkipLong(t)

	tests := []struct {
		name           string
		manager        func(*MockManager) *MockManager
		evaluator      func(*MockEvaluator) *MockEvaluator
		transformer    func(*MockTransformer) *MockTransformer
		leaderelection func(*uniqueness.MockManager) *uniqueness.MockManager
		resources      int
		expectedEvents []int
	}{
		{
			name: "Should return multiple results",
			manager: func(m *MockManager) *MockManager {
				m.EXPECT().Run().Return().Once()
				m.EXPECT().Stop().Once()
				return m
			},
			evaluator: func(m *MockEvaluator) *MockEvaluator {
				m.EXPECT().Eval(mock.Anything, mock.Anything).Return(evaluator.EventData{}, nil).Once()
				m.EXPECT().Eval(mock.Anything, mock.Anything).Return(evaluator.EventData{}, nil).Once()
				m.EXPECT().Eval(mock.Anything, mock.Anything).Return(evaluator.EventData{}, nil).Once()
				return m
			},
			transformer: func(m *MockTransformer) *MockTransformer {
				m.EXPECT().CreateBeatEvents(mock.Anything, mock.Anything).Return([]beat.Event{{}, {}}, nil).Once()
				m.EXPECT().CreateBeatEvents(mock.Anything, mock.Anything).Return([]beat.Event{}, nil).Once()
				m.EXPECT().CreateBeatEvents(mock.Anything, mock.Anything).Return([]beat.Event{{}, {}, {}}, nil).Once()
				return m
			},
			leaderelection: func(m *uniqueness.MockManager) *uniqueness.MockManager {
				m.EXPECT().Run(mock.Anything).Return(nil).Once()
				m.EXPECT().Stop().Once()
				return m
			},
			resources:      3,
			expectedEvents: []int{2, 0, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			sut := &k8sbenchmark{
				basebenchmark: basebenchmark{
					log:         testhelper.NewLogger(t),
					manager:     tt.manager(NewMockManager(t)),
					evaluator:   tt.evaluator(NewMockEvaluator(t)),
					transformer: tt.transformer(NewMockTransformer(t)),
					resourceCh:  make(chan fetching.ResourceInfo),
				},
				leaderElector: tt.leaderelection(uniqueness.NewMockManager(t)),
			}

			eventsCh, err := sut.Run(t.Context())
			require.NoError(t, err)
			for i := 0; i < tt.resources; i++ {
				sut.resourceCh <- fetching.ResourceInfo{} //nolint:exhaustruct
			}

			time.Sleep(100 * time.Millisecond)
			sut.Stop()

			for i, d := range tt.expectedEvents {
				events, ok := <-eventsCh
				assert.True(t, ok)
				assert.Len(t, events, d, "test %d", i)
			}

			_, ok := <-eventsCh
			assert.False(t, ok)
		})
	}
}
