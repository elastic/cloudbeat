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

package awsfetcher

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloudbeat/internal/inventory"
)

func collectResourcesAndMatch(t *testing.T, fetcher inventory.AssetFetcher, expected []inventory.AssetEvent) {
	t.Helper()

	ch := make(chan inventory.AssetEvent)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	go func() {
		fetcher.Fetch(ctx, ch)
	}()

	received := make([]inventory.AssetEvent, 0, len(expected))
	for len(expected) != len(received) {
		select {
		case <-ctx.Done():
			assert.ElementsMatch(t, expected, received)
			return
		case event := <-ch:
			received = append(received, event)
		}
	}

	assert.ElementsMatch(t, expected, received)
}
