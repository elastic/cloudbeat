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
	"github.com/stretchr/testify/mock"
	"go.uber.org/goleak"

	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

func TestPublisher_HandleEvents(t *testing.T) {
	testCases := []struct {
		name              string
		interval          time.Duration
		threshold         int
		ctxTimeout        time.Duration
		eventCount        int
		expectedEventSize []int
		closeChannel      bool
	}{
		{
			name:              "Publish events on threshold reached",
			interval:          time.Minute,
			threshold:         5,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        5,
			expectedEventSize: []int{5},
		},
		{
			name:              "Publish events on threshold reached twice",
			interval:          time.Minute,
			threshold:         5,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        10,
			expectedEventSize: []int{5, 5},
		},
		{
			name:              "Publish events on threshold reached twice and interval reached",
			interval:          100 * time.Millisecond,
			threshold:         5,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        13,
			expectedEventSize: []int{5, 5, 3},
		},
		{
			name:              "Publish events on threshold reached twice and context reached",
			interval:          time.Minute,
			threshold:         5,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        13,
			expectedEventSize: []int{5, 5, 3},
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
			eventCount:        5,
			expectedEventSize: []int{5},
		},
		{
			name:              "Publish events on interval reached 2 times",
			interval:          45 * time.Millisecond,
			threshold:         10,
			ctxTimeout:        200 * time.Millisecond,
			eventCount:        10,
			expectedEventSize: []int{5, 4, 1},
		},
		{
			name:              "Publish events on closed channel",
			interval:          time.Minute,
			threshold:         10,
			ctxTimeout:        time.Minute,
			eventCount:        5,
			expectedEventSize: []int{5},
			closeChannel:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			log := testhelper.NewLogger(t)
			goleak.VerifyNone(t, goleak.IgnoreCurrent())
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()

			client := newMockClient(t)
			for _, size := range tc.expectedEventSize {
				client.EXPECT().PublishAll(mock.MatchedBy(LengthMatcher(size)))
			}
			publisher := NewPublisher(log, tc.interval, tc.threshold, client)

			eventsChannel := make(chan beat.Event)

			go func() {
				for i := 0; i < tc.eventCount; i++ {
					select {
					case <-ctx.Done():
						return
					// Simulate events being sent to the channel
					case eventsChannel <- beat.Event{}:
					}
					time.Sleep(10 * time.Millisecond)
				}
				if tc.closeChannel {
					close(eventsChannel)
				}
			}()

			publisher.HandleEvents(ctx, eventsChannel)
		})
	}
}

func LengthMatcher(length int) func(events []beat.Event) bool {
	return func(events []beat.Event) bool {
		return len(events) == length
	}
}
