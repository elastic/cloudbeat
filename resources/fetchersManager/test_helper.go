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

package fetchersManager

import (
	"context"
	"github.com/elastic/cloudbeat/resources/fetchersManager/factory"
	//"github.com/elastic/cloudbeat/resources/fetchersManager/registry"
	"github.com/elastic/cloudbeat/resources/fetching"
	"sync"
)

type NumberFetcher struct {
	num        int
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
	wg         *sync.WaitGroup
}
type syncNumberFetcher struct {
	num        int
	stopCalled bool
	resourceCh chan fetching.ResourceInfo
}

func (f *syncNumberFetcher) Fetch(_ context.Context, cMetadata fetching.CycleMetadata) error {
	f.resourceCh <- fetching.ResourceInfo{
		Resource:      FetchValue(f.num),
		CycleMetadata: cMetadata,
	}

	return nil
}

func (f *syncNumberFetcher) Stop() {
	f.stopCalled = true
}

func NewSyncNumberFetcher(num int, ch chan fetching.ResourceInfo) fetching.Fetcher {
	return &syncNumberFetcher{num, false, ch}
}

type NumberResource struct {
	Num int
}

func (res NumberResource) GetData() interface{} {
	return res.Num
}

func (res NumberResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      "",
		Type:    "number",
		SubType: "number",
		Name:    "number",
	}, nil
}

func (res NumberResource) GetElasticCommonData() interface{} {
	return nil
}

//func RegisterNFetchers(t *testing.T, reg registry.FetchersRegistry, n int, ch chan fetching.ResourceInfo, wg *sync.WaitGroup) {
//	fMap := make(factory.FetchersMap, 0)
//	for i := 0; i < n; i++ {
//		key := fmt.Sprint(i)
//		fMap[key] = factory.RegisteredFetcher{
//			Fetcher:   NewNumberFetcher(i, ch, wg),
//			Condition: nil,
//		}
//	}
//}

func RegisterFetcher(fMap factory.FetchersMap, f fetching.Fetcher, key string, condition []fetching.Condition) {
	fMap[key] = factory.RegisteredFetcher{Fetcher: f, Condition: condition}
}

func NewNumberFetcher(num int, ch chan fetching.ResourceInfo, wg *sync.WaitGroup) fetching.Fetcher {
	return &NumberFetcher{num, false, ch, wg}
}

func (f *NumberFetcher) Fetch(_ context.Context, cMetadata fetching.CycleMetadata) error {
	defer f.wg.Done()

	f.resourceCh <- fetching.ResourceInfo{
		Resource:      FetchValue(f.num),
		CycleMetadata: cMetadata,
	}

	return nil
}

func (f *NumberFetcher) Stop() {
	f.stopCalled = true
}

func FetchValue(num int) fetching.Resource {
	return NumberResource{num}
}

type boolFetcherCondition struct {
	val  bool
	name string
}

func NewBoolFetcherCondition(val bool, name string) fetching.Condition {
	return &boolFetcherCondition{val, name}
}

func (c *boolFetcherCondition) Condition() bool {
	return c.val
}

func (c *boolFetcherCondition) Name() string {
	return c.name
}
