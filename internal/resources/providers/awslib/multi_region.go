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
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

type RegionsSelector interface {
	Regions(ctx context.Context, cfg aws.Config) ([]string, error)
}

type CrossRegionFetcher[T any] interface {
	GetMultiRegionsClientMap() map[string]T
}

type CrossRegionFactory[T any] interface {
	NewMultiRegionClients(ctx context.Context, selector RegionsSelector, cfg aws.Config, factory func(cfg aws.Config) T, log *clog.Logger) CrossRegionFetcher[T]
}

type (
	MultiRegionClientFactory[T any] struct{}
	multiRegionWrapper[T any]       struct {
		clients map[string]T
	}
)

// NewMultiRegionClients is a utility function that is used to create a map of client instances of a given type T for multiple regions.
func (w *MultiRegionClientFactory[T]) NewMultiRegionClients(ctx context.Context, selector RegionsSelector, cfg aws.Config, factory func(cfg aws.Config) T, log *clog.Logger) CrossRegionFetcher[T] {
	clientsMap := make(map[string]T, 0)
	regionList, err := selector.Regions(ctx, cfg)
	if err != nil {
		log.Errorf("Region '%s' selected after failure to retrieve aws regions: %v", cfg.Region, err)
		regionList = []string{cfg.Region}
	}
	for _, region := range regionList {
		cfg.Region = region
		clientsMap[region] = factory(cfg)
	}

	wrapper := &multiRegionWrapper[T]{
		clients: clientsMap,
	}

	return wrapper
}

// MultiRegionFetch retrieves resources from multiple regions concurrently using the provided fetcher function.
func MultiRegionFetch[T any, K any](ctx context.Context, set map[string]T, fetcher func(ctx context.Context, region string, client T) (K, error)) ([]K, error) {
	var wg sync.WaitGroup
	var mux sync.Mutex
	var crossRegionResources []K
	errChan := make(chan error, len(set))

	if set == nil {
		return nil, errors.New("multi region clients have not been initialized")
	}

	for region, client := range set {
		wg.Add(1)
		go func(client T, region string, errCn chan error) {
			defer wg.Done()
			results, fetchErr := fetcher(ctx, region, client)
			if fetchErr != nil {
				errCn <- errors.Join(fmt.Errorf("fail to retrieve aws resources for region: %s", region), fetchErr)
			}

			mux.Lock()
			defer mux.Unlock()
			// K might be an slice, struct or pointer
			// in case of pointer we do not want to return a slice with nils
			if shouldDrop(results) {
				return
			}
			crossRegionResources = append(crossRegionResources, results)
		}(client, region, errChan)
	}

	wg.Wait()
	select {
	case err := <-errChan:
		return crossRegionResources, err
	default:
		return crossRegionResources, nil
	}
}

// shouldDrop checks the target type and return true if
// the type is pointer -> the pointer is nil
// and false otherwise
func shouldDrop(t any) bool {
	v := reflect.ValueOf(t)
	kind := v.Kind()
	if kind == reflect.Ptr && v.IsNil() {
		return true
	}

	// shouldDrop(nil) case
	if kind == reflect.Invalid && t == nil {
		return true
	}

	return false
}

func (w *multiRegionWrapper[T]) GetMultiRegionsClientMap() map[string]T {
	return w.clients
}
