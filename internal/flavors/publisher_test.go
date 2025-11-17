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
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestPublisher_HandleEvents(t *testing.T) {
	testhelper.SkipLong(t)

	type testCase struct {
		name                  string
		interval              time.Duration
		threshold             int
		ctxTimeout            time.Duration
		eventCount            int
		expectedEventSize     []int
		closeChannel          bool
		expectedEventsInCycle int
	}
	testCases := []testCase{
		{
			name:              "Publish single event on threshold reached",
			interval:          time.Minute,
			threshold:         1,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        2,
			expectedEventSize: []int{1},
		},
		{
			name:              "Publish events on threshold reached",
			interval:          time.Minute,
			threshold:         10,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        5,
			expectedEventSize: []int{10},
		},
		{
			name:              "Publish events on threshold reached twice",
			interval:          time.Minute,
			threshold:         5,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        6,
			expectedEventSize: []int{6, 9},
		},
		{
			name:              "Publish events on threshold reached twice and interval reached",
			interval:          100 * time.Millisecond,
			threshold:         10,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        8,
			expectedEventSize: []int{10, 11, 7},
		},
		{
			name:              "Publish events on threshold reached twice and channel closed",
			interval:          time.Minute,
			threshold:         10,
			ctxTimeout:        time.Minute,
			eventCount:        8,
			expectedEventSize: []int{10, 11, 7},
			closeChannel:      true,
		},
		{
			name:              "Publish events on threshold reached twice and context reached",
			interval:          time.Minute,
			threshold:         10,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        8,
			expectedEventSize: []int{10, 11, 7},
		},
		{
			name:              "Publish events on context done",
			interval:          time.Minute,
			threshold:         5,
			ctxTimeout:        100 * time.Millisecond,
			eventCount:        3,
			expectedEventSize: []int{3},
		},
		{
			name:              "Publish 0 events on context done",
			interval:          time.Minute,
			threshold:         5,
			ctxTimeout:        0,
			eventCount:        5,
			expectedEventSize: []int{},
		},
		{
			name:              "Publish events on interval reached",
			interval:          100 * time.Millisecond,
			threshold:         10,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        4,
			expectedEventSize: []int{6},
		},
		{
			name:              "Publish events on closed channel",
			interval:          time.Minute,
			threshold:         10,
			ctxTimeout:        time.Minute,
			eventCount:        4,
			expectedEventSize: []int{6},
			closeChannel:      true,
		},
		{
			name:                  "Publish events on interval reached 2 times",
			interval:              55 * time.Millisecond,
			threshold:             100,
			ctxTimeout:            250 * time.Millisecond,
			eventCount:            10,
			expectedEventSize:     []int{10, 26, 9},
			expectedEventsInCycle: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log := testhelper.NewObserverLogger(t)
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
			ctx, cancel := context.WithTimeout(t.Context(), tc.ctxTimeout)
			defer cancel()

			client := newMockClient(t)
			for _, size := range tc.expectedEventSize {
				client.EXPECT().PublishAll(mock.MatchedBy(lengthMatcher(size)))
			}
			publisher := NewPublisher(log, tc.interval, tc.threshold, client)

			eventsChannel := make(chan []beat.Event)

			go func(tc testCase) {
				start := time.Now()
				for i := 0; i < tc.eventCount; i++ {
					select {
					case <-ctx.Done():
						return
					// Simulate events being sent to the channel
					case eventsChannel <- generateEvents(t, i, tc.expectedEventsInCycle, tc.interval, start):
					}
					time.Sleep(10 * time.Millisecond)
				}
				if tc.closeChannel {
					close(eventsChannel)
				}
			}(tc)

			publisher.HandleEvents(ctx, eventsChannel)
			logs := logp.ObserverLogs().FilterFieldKey(ecsEventActionField).All()
			assert.Len(t, logs, len(tc.expectedEventSize))
			for i, size := range tc.expectedEventSize {
				assert.Equal(t, int64(size), logs[i].ContextMap()[ecsEventCountField])
			}
		})
	}
}

func lengthMatcher(length int) func(events []beat.Event) bool {
	return func(events []beat.Event) bool {
		return len(events) == length
	}
}

func generateEvents(t *testing.T, size int, expectedEventsInCycle int, interval time.Duration, start time.Time) []beat.Event {
	sleepOnCycleEnd(t, size, expectedEventsInCycle, interval, start)
	t.Logf(" %d events at %dms", size, time.Since(start).Milliseconds())
	results := make([]beat.Event, size)
	for i := range size {
		results[i] = beat.Event{}
	}
	return results
}

const gracePeriod = 10 * time.Millisecond

// sleepOnCycleEnd once the expected amount of events to be published is reached,
// wait the rest of the interval + grace period. The grace periods exists to avoid delays on the tick
func sleepOnCycleEnd(t *testing.T, size int, eventCount int, interval time.Duration, start time.Time) {
	if eventCount > 0 && size > 1 && size%eventCount == 1 {
		cycle := size / eventCount
		cycleInterval := interval * time.Duration(cycle)
		waitPeriod := cycleInterval - time.Since(start) + gracePeriod
		t.Logf("--- Waiting %s (cycle %d interval %s)", waitPeriod.String(), cycle, cycleInterval.String())
		time.Sleep(waitPeriod)
	}
}

func TestPublisher_HandleEvents_Buffer(t *testing.T) {
	testhelper.SkipLong(t)

	event := func(id int) beat.Event {
		return beat.Event{
			Fields: mapstr.M{"id": id},
		}
	}

	events := func(ids []int) []beat.Event {
		return lo.Map(ids, func(id int, _ int) beat.Event { return event(id) })
	}

	eventsBatches := func(batchesOfSize ...int) [][]beat.Event {
		id := 1
		return lo.Map(batchesOfSize, func(singleBatchSize int, _ int) []beat.Event {
			b := events(lo.RangeFrom(id, singleBatchSize))
			id += singleBatchSize
			return b
		})
	}

	tests := map[string]struct {
		interval                            time.Duration
		threshold                           int
		incomeEventBatchesSizes             []int
		expectedPublishedBatchesIDs         [][]int
		expectedBufferCapacityOnEachPublish []int
		expectedBufferCapacityEnd           int
		ctxTimeout                          time.Duration
	}{
		"single batch": {
			interval:                            2 * time.Second,
			threshold:                           10,
			incomeEventBatchesSizes:             []int{5},
			expectedPublishedBatchesIDs:         [][]int{{1, 2, 3, 4, 5}},
			expectedBufferCapacityOnEachPublish: []int{15},
			expectedBufferCapacityEnd:           15,
		},
		"single batch with ctx deadline": {
			interval:                            500 * time.Millisecond,
			threshold:                           10,
			incomeEventBatchesSizes:             []int{5},
			expectedPublishedBatchesIDs:         [][]int{{1, 2, 3, 4, 5}},
			expectedBufferCapacityOnEachPublish: []int{15},
			expectedBufferCapacityEnd:           15,
			ctxTimeout:                          time.Second,
		},
		"single batch increase capacity": {
			interval:                            2 * time.Second,
			threshold:                           5,
			incomeEventBatchesSizes:             []int{10},
			expectedPublishedBatchesIDs:         [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
			expectedBufferCapacityOnEachPublish: []int{15},
			expectedBufferCapacityEnd:           15,
		},
		"two batches under threshold": {
			interval:                            2 * time.Second,
			threshold:                           10,
			incomeEventBatchesSizes:             []int{5, 5},
			expectedPublishedBatchesIDs:         [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
			expectedBufferCapacityOnEachPublish: []int{15},
			expectedBufferCapacityEnd:           15,
		},
		"two batches at threshold": {
			interval:                            2 * time.Second,
			threshold:                           5,
			incomeEventBatchesSizes:             []int{3, 3, 6},
			expectedPublishedBatchesIDs:         [][]int{{1, 2, 3, 4, 5, 6}, {7, 8, 9, 10, 11, 12}},
			expectedBufferCapacityOnEachPublish: []int{7, 7},
			expectedBufferCapacityEnd:           7,
		},
		"single batch over threshold": {
			interval:                            2 * time.Second,
			threshold:                           3,
			incomeEventBatchesSizes:             []int{5},
			expectedPublishedBatchesIDs:         [][]int{{1, 2, 3, 4, 5}},
			expectedBufferCapacityOnEachPublish: []int{8},
			expectedBufferCapacityEnd:           8, // 4 * 2
		},
		"single batch over threshold break capacity limit": {
			interval:                            2 * time.Second,
			threshold:                           3,
			incomeEventBatchesSizes:             []int{15},
			expectedPublishedBatchesIDs:         [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}},
			expectedBufferCapacityOnEachPublish: []int{15},
			expectedBufferCapacityEnd:           4,
		},
		"multiple batches break capacity limit": {
			interval:                            2 * time.Second,
			threshold:                           5,
			incomeEventBatchesSizes:             []int{2, 2, 20},
			expectedPublishedBatchesIDs:         [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24}},
			expectedBufferCapacityOnEachPublish: []int{27},
			expectedBufferCapacityEnd:           7,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			log := testhelper.NewLogger(t)
			defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.ctxTimeout != 0 {
				ctx, cancel = context.WithTimeout(ctx, tc.ctxTimeout)
				defer cancel()
			}

			var publisher *Publisher

			client := newMockClient(t)
			calls := make([]*mock.Call, len(tc.expectedPublishedBatchesIDs))
			for i, publishedSliceIDs := range tc.expectedPublishedBatchesIDs {
				calls[i] = client.EXPECT().PublishAll(events(publishedSliceIDs)).Run(
					func(_ []beat.Event) {
						assert.Equalf(t, tc.expectedBufferCapacityOnEachPublish[i], cap(publisher.eventsBuffer), "buffer capacity miss-match on publish number %d", i)
					},
				).Call
			}
			mock.InOrder(calls...)

			publisher = NewPublisher(log, tc.interval, tc.threshold, client)
			eventsChannel := make(chan []beat.Event, 10)

			// send events
			go func() {
				events := eventsBatches(tc.incomeEventBatchesSizes...)
				for _, batch := range events {
					eventsChannel <- batch
				}
				if tc.ctxTimeout == 0 {
					close(eventsChannel)
				}
			}()

			publisher.HandleEvents(ctx, eventsChannel)
			require.Equal(t, tc.expectedBufferCapacityEnd, cap(publisher.eventsBuffer))
		})
	}
}
