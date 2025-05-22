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

package cycle

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type helper struct {
	value  int
	err    error
	called atomic.Int64
}

func (h *helper) cb(_ context.Context) (int, error) {
	h.called.Add(1)
	return h.value, h.err
}

func TestCache(t *testing.T) {
	h := helper{}
	cache := NewCache[int](testhelper.NewLogger(t))

	tests := []struct {
		name string

		sequence int64
		setValue int
		setError bool

		want       int
		wantErr    bool
		wantCalled bool
	}{
		{
			name:       "error first",
			sequence:   10,
			setError:   true,
			want:       0,
			wantErr:    true,
			wantCalled: true,
		},
		{
			name:       "called with error again",
			sequence:   0,
			setError:   true,
			want:       0,
			wantErr:    true,
			wantCalled: true,
		},
		{
			name:       "success",
			sequence:   0,
			setValue:   1,
			setError:   false,
			want:       1,
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "success call skipped",
			sequence:   0,
			setValue:   100,  // not used
			setError:   true, // not used
			want:       1,    // previous value
			wantErr:    false,
			wantCalled: false,
		},
		{
			name:       "new cycle error",
			sequence:   1,
			setError:   true,
			want:       1, // previous value
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "same cycle but with success",
			sequence:   1,
			setValue:   2,
			setError:   false,
			want:       2,
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "old cycle",
			sequence:   0,
			want:       2, // previous value
			wantErr:    false,
			wantCalled: false,
		},
		{
			name:       "new cycle",
			sequence:   5,
			setValue:   3,
			setError:   false,
			want:       3,
			wantErr:    false,
			wantCalled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.value = tt.setValue
			if tt.setError {
				h.err = errors.New("some error")
			} else {
				h.err = nil
			}
			h.called.Store(0)
			got, err := cache.GetValue(t.Context(), Metadata{Sequence: tt.sequence}, h.cb)
			if tt.wantErr {
				require.ErrorContains(t, err, "some error")
			} else {
				require.NoError(t, err)
			}
			var expectedCalls int64
			if tt.wantCalled {
				expectedCalls = 1
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, expectedCalls, h.called.Load())
		})
	}
}

func TestCache_Lock(t *testing.T) {
	testhelper.SkipLong(t)

	ctx := t.Context()
	count := 0
	ch := make(chan struct{})
	fetch := func(_ context.Context) (int, error) {
		count++
		if count == 1 {
			close(ch)
			time.Sleep(50 * time.Millisecond)
			return 1, nil
		}
		return -1, errors.New("some error")
	}
	cache := NewCache[int](testhelper.NewLogger(t))

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Wait() // wait for tests in goroutine
	go func() {
		defer wg.Done()

		got, err := cache.GetValue(ctx, Metadata{Sequence: 1}, fetch)
		assert.NoError(t, err)
		assert.Equal(t, 1, got)
	}()

	<-ch // wait until callback is blocked
	got, err := cache.GetValue(ctx, Metadata{Sequence: 1}, fetch)
	require.NoError(t, err)
	assert.Equal(t, 1, got)
	assert.Equal(t, 1, count)
}
