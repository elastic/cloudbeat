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

package testhelper

import (
	"testing"
	"time"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

// CollectResources fetches items from a channel and returns them in a slice.
//
// Warning: this function does not wait for the channel to close, using it can cause race conditions.
// See CollectResourcesBlocking for a blocking version of the function.
func CollectResources[T any](ch chan T) []T {
	var results []T
	for {
		select {
		case value, ok := <-ch:
			if !ok {
				return results
			}
			results = append(results, value)
		default:
			return results
		}
	}
}

// CollectResourcesWithTimeout fetches items from a channel and returns them in a slice after no elements have been
// received for the specified timeout duration.
func CollectResourcesWithTimeout[T any](ch chan T, maxCount int, timeout time.Duration) []T {
	var results []T
	for {
		if len(results) >= maxCount {
			return results
		}

		select {
		case value, ok := <-ch:
			if !ok {
				return results
			}
			results = append(results, value)
		case <-time.After(timeout):
			return results
		}
	}
}

// CollectResourcesBlocking fetches items from a channel and returns them in a slice.
// This function waits for the channel to close before returning.
// See CollectResources for a non-blocking version of the function.
func CollectResourcesBlocking[T any](ch chan T) []T {
	results := make([]T, 0)
	for v := range ch {
		results = append(results, v)
	}
	return results
}

func CreateMockClients[T any](client T, regions []string) map[string]T {
	m := make(map[string]T, 0)
	for _, clientRegion := range regions {
		m[clientRegion] = client
	}

	return m
}

func NewLogger(t *testing.T) *clog.Logger {
	t.Helper()

	require.NoError(t, logp.TestingSetup())

	return clog.NewLogger(t.Name())
}

func NewObserverLogger(t *testing.T) *clog.Logger {
	t.Helper()

	require.NoError(t, logp.DevelopmentSetup(logp.ToObserverOutput()))

	return clog.NewLogger(t.Name())
}

func SkipLong(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
}
