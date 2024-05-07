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
)

type MapCache[T any] struct {
	results sync.Map
}

func (c *MapCache[T]) Get(fn func() T, key string) T {
	if value, ok := c.results.Load(key); ok {
		return value.(T)
	}

	value := fn()
	c.results.Store(key, value)
	return value
}

func NewMapCache[T any]() *MapCache[T] {
	return &MapCache[T]{}
}
