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

package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloudbeat/internal/inventory"
)

func CollectResourcesAndMatch(t *testing.T, fetcher inventory.AssetFetcher, expected []inventory.AssetEvent) {
	t.Helper()

	ch := make(chan inventory.AssetEvent)
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	go func() {
		defer close(ch)
		fetcher.Fetch(ctx, ch)
	}()

	received := make([]inventory.AssetEvent, 0, len(expected))
	defer func() {
		assert.ElementsMatch(t, expected, received)
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			received = append(received, event)
		}
	}
}
