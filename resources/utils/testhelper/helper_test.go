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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCollectResources(t *testing.T) {
	t.Run("empty channel", func(t *testing.T) {
		ch := make(chan struct{})
		assert.Empty(t, CollectResources(ch))
	})
	t.Run("one item", func(t *testing.T) {
		ch := make(chan struct{})
		write(ch)
		assert.Len(t, CollectResources(ch), 1)
	})
	t.Run("one item later", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		ch := make(chan struct{})

		time.AfterFunc(100*time.Millisecond, func() {
			defer wg.Done()
			write(ch)
			assert.Len(t, CollectResources(ch), 1)
		})

		write(ch)
		assert.Len(t, CollectResources(ch), 1)
		assert.Empty(t, CollectResources(ch))

		wg.Wait()
	})
	t.Run("buffered channel", func(t *testing.T) {
		ch := make(chan struct{}, 10)
		ch <- struct{}{}
		assert.Len(t, CollectResources(ch), 1)
	})
	t.Run("closed channel", func(t *testing.T) {
		ch := make(chan struct{})
		close(ch)
		assert.Empty(t, CollectResources(ch))
	})
}

func TestCollectResourcesBlocking(t *testing.T) {
	t.Run("blocks on empty channel", func(t *testing.T) {
		ch := make(chan struct{})
		time.AfterFunc(10*time.Millisecond, func() {
			ch <- struct{}{}
			ch <- struct{}{}
			ch <- struct{}{}
			close(ch)
		})
		assert.Len(t, CollectResourcesBlocking(ch), 3)
	})
	t.Run("closed channel", func(t *testing.T) {
		ch := make(chan struct{})
		close(ch)
		assert.Empty(t, CollectResourcesBlocking(ch))
	})
}

func TestCollectResourcesWithTimeout(t *testing.T) {
	t.Run("blocks on empty channel", func(t *testing.T) {
		ch := make(chan struct{})

		time.AfterFunc(100*time.Millisecond, func() {
			ch <- struct{}{}
			ch <- struct{}{}
			ch <- struct{}{}
		})

		assert.Empty(t, CollectResourcesWithTimeout(ch, 100, 10*time.Millisecond))
		assert.Len(t, CollectResourcesWithTimeout(ch, 100, 100*time.Millisecond), 3)
	})
	t.Run("closed channel", func(t *testing.T) {
		done := make(chan struct{})

		go func() {
			ch := make(chan struct{})
			close(ch)
			assert.Empty(t, CollectResourcesWithTimeout(ch, 100, 10*time.Hour))
			done <- struct{}{}
		}()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("time out")
		}
	})
	t.Run("max count", func(t *testing.T) {
		ch := make(chan struct{})
		assert.Empty(t, CollectResourcesWithTimeout(ch, 0, time.Hour))

		time.AfterFunc(10*time.Millisecond, func() {
			ch <- struct{}{}
			ch <- struct{}{}
			ch <- struct{}{}
		})

		time.AfterFunc(100*time.Millisecond, func() {
			ch <- struct{}{}
			ch <- struct{}{}
			ch <- struct{}{}
		})

		assert.Len(t, CollectResourcesWithTimeout(ch, 2, 20*time.Millisecond), 2)
		assert.Len(t, CollectResourcesWithTimeout(ch, 4, 200*time.Millisecond), 1+3)
		assert.Empty(t, CollectResourcesWithTimeout(ch, 100, 10*time.Millisecond))
	})
}

func write(ch chan struct{}) {
	go func() {
		ch <- struct{}{}
	}()
	time.Sleep(10 * time.Millisecond)
}
