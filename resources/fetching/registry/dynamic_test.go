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

package registry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
)

type testData struct {
	lock        sync.Mutex
	m           factory.FetchersMap
	updateError error
}

func (td *testData) updateFunc() (factory.FetchersMap, error) {
	td.lock.Lock()
	defer td.lock.Unlock()

	ret := make(factory.FetchersMap)
	for k, v := range td.m {
		ret[k] = v
	}
	return ret, td.updateError
}

func (td *testData) set(key string, f fetching.Fetcher) {
	td.lock.Lock()
	defer td.lock.Unlock()

	if td.m == nil {
		td.m = make(factory.FetchersMap)
	}

	td.m[key] = factory.RegisteredFetcher{Fetcher: f}
}

func (td *testData) empty() {
	td.lock.Lock()
	defer td.lock.Unlock()

	for k := range td.m {
		delete(td.m, k)
	}
}

func (td *testData) setError(err error) {
	td.lock.Lock()
	defer td.lock.Unlock()

	td.updateError = err
}

func TestDynamic(t *testing.T) {
	period := 50 * time.Millisecond

	td := testData{}
	d := NewDynamic(testhelper.NewLogger(t), period, td.updateFunc)

	t.Run("empty registry", func(t *testing.T) {
		assert.False(t, d.ShouldRun("some-key"))
		assert.Empty(t, d.Keys())
	})

	t.Run("add fetchers", func(t *testing.T) {
		f := fetching.NewMockFetcher(t)
		f.EXPECT().Fetch(mock.Anything, mock.Anything).Once().Return(nil)
		f.EXPECT().Stop().Once()
		td.set("some-key", f)

		f = fetching.NewMockFetcher(t)
		f.EXPECT().Fetch(mock.Anything, mock.Anything).Once().Return(errors.New("some-error"))
		f.EXPECT().Stop().Once()
		td.set("fetcher-with-error", f)

		defer d.Stop()

		time.Sleep(2 * period)
		assert.True(t, d.ShouldRun("some-key"))
		assert.True(t, d.ShouldRun("fetcher-with-error"))
		assert.False(t, d.ShouldRun("some-other-key"))
		assert.Len(t, d.Keys(), 2)

		assert.NoError(t, d.Run(context.Background(), "some-key", fetching.CycleMetadata{}))
		assert.ErrorContains(t, d.Run(context.Background(), "fetcher-with-error", fetching.CycleMetadata{}), "some-error")

		td.empty()

		t.Run("error is ignored", func(t *testing.T) {
			td.setError(errors.New("update error"))

			time.Sleep(2 * period)
			assert.Len(t, d.Keys(), 2)
		})
	})

	t.Run("empty again", func(t *testing.T) {
		defer d.Stop()

		td.setError(nil)

		time.Sleep(2 * period)
		assert.False(t, d.ShouldRun("some-key"))
		assert.Empty(t, d.Keys())
	})

	t.Run("stop again", func(t *testing.T) {
		d.Stop()
		d.Stop()
		d.Stop()
		assert.Nil(t, d.(*dynamic).timer)
	})
}
