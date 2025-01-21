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

package awslib

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dgraph-io/ristretto"

	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

var (
	ristrettoCache        *ristretto.Cache[string, *cachedRegions]
	allRegionCacheTTL     = 720 * time.Hour
	currentRegionCacheTTL time.Duration
)

func init() {
	var err error
	ristrettoCache, err = newCachedRegions()
	if err != nil {
		panic(fmt.Errorf("unable to init region-selector cache: %w", err))
	}
}

func newCachedRegions() (*ristretto.Cache[string, *cachedRegions], error) {
	return ristretto.NewCache(&ristretto.Config[string, *cachedRegions]{
		NumCounters: 100,
		MaxCost:     10000,
		BufferItems: 64,
	})
}

func CurrentRegionSelector() RegionsSelector {
	return newCachedRegionSelector(&currentRegionSelector{}, "CurrentRegionSelectorCache", currentRegionCacheTTL)
}

func AllRegionSelector() RegionsSelector {
	return newCachedRegionSelector(&allRegionSelector{}, "AllRegionSelectorCache", allRegionCacheTTL)
}

type cachedRegions struct {
	regions []string
}

type cachedRegionSelector struct {
	lock   *sync.RWMutex
	keep   time.Duration
	key    string
	client RegionsSelector
}

func newCachedRegionSelector(selector RegionsSelector, cache string, keep time.Duration) *cachedRegionSelector {
	return &cachedRegionSelector{
		lock:   &sync.RWMutex{},
		keep:   keep,
		key:    cache,
		client: selector,
	}
}

func (s *cachedRegionSelector) Regions(ctx context.Context, cfg aws.Config) ([]string, error) {
	log := clog.NewLogger("aws")

	// Make sure that consequent calls to the function will keep trying to retrieve the regions list until it succeeds.
	cachedObject := s.getCache()
	if cachedObject != nil {
		return cachedObject, nil
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	log.Debug("RegionsSelector starting to retrieve regions")
	var output []string
	output, err := s.client.Regions(ctx, cfg)
	if err != nil {
		log.Errorf("Failed getting regions: %v", err)
		return nil, err
	}

	if !s.setCache(output) {
		log.Errorf("Failed setting regions cache")
	}
	return output, nil
}

func (s *cachedRegionSelector) setCache(list []string) bool {
	cache := &cachedRegions{
		regions: list,
	}

	return ristrettoCache.SetWithTTL(s.key, cache, 1, s.keep)
}

func (s *cachedRegionSelector) getCache() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	cachedObject, ok := ristrettoCache.Get(s.key)
	if !ok {
		return nil
	}

	return cachedObject.regions
}
