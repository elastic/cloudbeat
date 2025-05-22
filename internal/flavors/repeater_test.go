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

package flavors

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestRepeater_Run(t *testing.T) {
	testhelper.SkipLong(t)

	testCases := []struct {
		name           string
		interval       time.Duration
		ctxTimeout     time.Duration
		fnMock         func(*testing.T) *MockRepeaterFunc
		expectedErrMsg string
	}{
		{
			name:       "Function succeeds once",
			interval:   100 * time.Millisecond,
			ctxTimeout: 20 * time.Millisecond,
			fnMock: func(t *testing.T) *MockRepeaterFunc {
				m := NewMockRepeaterFunc(t)
				m.EXPECT().Execute().Return(nil).Once()
				return m
			},
			expectedErrMsg: "",
		},
		{
			name:       "Function succeeds 4 times",
			interval:   100 * time.Millisecond,
			ctxTimeout: 350 * time.Millisecond,
			fnMock: func(t *testing.T) *MockRepeaterFunc {
				m := NewMockRepeaterFunc(t)
				m.EXPECT().Execute().Return(nil).Times(4)
				return m
			},
			expectedErrMsg: "",
		},
		{
			name:       "Function delays in less than interval",
			interval:   80 * time.Millisecond,
			ctxTimeout: 320 * time.Millisecond,
			fnMock: func(t *testing.T) *MockRepeaterFunc {
				m := NewMockRepeaterFunc(t)
				m.EXPECT().Execute().After(100 * time.Millisecond).Return(nil).Times(4)
				m.EXPECT().Execute().Maybe().Return(nil)
				return m
			},
			expectedErrMsg: "",
		},
		{
			name:       "Function delays in more than interval",
			interval:   100 * time.Millisecond,
			ctxTimeout: 350 * time.Millisecond,
			fnMock: func(t *testing.T) *MockRepeaterFunc {
				m := NewMockRepeaterFunc(t)
				m.EXPECT().Execute().After(200 * time.Millisecond).Return(nil).Times(2)
				m.EXPECT().Execute().Maybe().Return(nil)
				return m
			},
			expectedErrMsg: "",
		},
		{
			name:       "Function returns error",
			interval:   100 * time.Millisecond,
			ctxTimeout: 500 * time.Millisecond,
			fnMock: func(t *testing.T) *MockRepeaterFunc {
				m := NewMockRepeaterFunc(t)
				m.EXPECT().Execute().Return(errors.New("test error")).Once()
				return m
			},
			expectedErrMsg: "test error",
		},
		{
			name:       "Context canceled",
			interval:   500 * time.Millisecond,
			ctxTimeout: 100 * time.Millisecond,
			fnMock: func(t *testing.T) *MockRepeaterFunc {
				m := NewMockRepeaterFunc(t)
				m.EXPECT().Execute().Return(nil).Once()
				return m
			},
			expectedErrMsg: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log := testhelper.NewLogger(t)
			repeater := NewRepeater(log, tc.interval)
			ctx, cancel := context.WithTimeout(t.Context(), tc.ctxTimeout)
			defer cancel()

			m := tc.fnMock(t)
			err := repeater.Run(ctx, m.Execute)

			if tc.expectedErrMsg != "" {
				require.ErrorContains(t, err, tc.expectedErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}
