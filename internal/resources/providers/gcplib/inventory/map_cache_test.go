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

package inventory

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFunction struct {
	mock.Mock
}

func (m *MockFunction) GetSomeValue() int {
	m.Called()
	return 0
}

func TestMapCacheGet(t *testing.T) {
	cache := NewMapCache[int]()

	// Test getting existing value from cache
	cache.results.Store("key1", 42)
	mockFunction := new(MockFunction)
	result := cache.Get(mockFunction.GetSomeValue, "key1")
	mockFunction.AssertNotCalled(t, "GetSomeValue")
	assert.Equal(t, 42, result)

	// Test getting non-existing value from cache
	mockFunction.On("GetSomeValue").Return(mockFunction.GetSomeValue())
	result = cache.Get(mockFunction.GetSomeValue, "key2")
	mockFunction.AssertNumberOfCalls(t, "GetSomeValue", 2) // 1 by Return(), 2nd by cache.Get()
	assert.Equal(t, 0, result)

	// Test concurrent accesses
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.Get(func() int { return 1 }, "concurrent_key")
		}()
	}
	wg.Wait()
}
