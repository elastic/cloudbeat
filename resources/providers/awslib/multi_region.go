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
	"sync"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elastic/elastic-agent-libs/logp"
)

var (
	awsRegions *cachedRegions
	once       = &sync.Once{}
)

type cachedRegions struct {
	enabledRegions []string
}

type CrossRegionFetcher[T any] interface {
	GetMultiRegionsClientMap() map[string]T
}

type CrossRegionFactory[T any] interface {
	NewMultiRegionClients(client DescribeCloudRegions, cfg awssdk.Config, factory func(cfg awssdk.Config) T, log *logp.Logger) CrossRegionFetcher[T]
}

type DescribeCloudRegions interface {
	DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)
}

type (
	MultiRegionClientFactory[T any] struct{}
	multiRegionWrapper[T any]       struct {
		clients map[string]T
	}
)

// NewMultiRegionClients is a utility function that is used to create a map of client instances of a given type T for multiple regions.
func (w *MultiRegionClientFactory[T]) NewMultiRegionClients(client DescribeCloudRegions, cfg awssdk.Config, factory func(cfg awssdk.Config) T, log *logp.Logger) CrossRegionFetcher[T] {
	clientsMap := make(map[string]T, 0)
	for _, region := range getRegions(client, log) {
		cfg.Region = region
		clientsMap[region] = factory(cfg)
	}

	wrapper := &multiRegionWrapper[T]{
		clients: clientsMap,
	}

	return wrapper
}

// Fetch retrieves resources from multiple regions concurrently using the provided fetcher function.
func MultiRegionFetch[T any, K any](ctx context.Context, set map[string]T, fetcher func(ctx context.Context, client T) (K, error)) ([]K, error) {
	var err error
	var wg sync.WaitGroup
	var mux sync.Mutex
	var crossRegionResources []K

	if set == nil {
		return nil, errors.New("multi region clients have not been initialize")
	}

	for region, client := range set {
		wg.Add(1)
		go func(client T, region string) {
			defer wg.Done()
			results, fetchErr := fetcher(ctx, client)
			if fetchErr != nil {
				err = fmt.Errorf("fail to retrieve aws resources for region: %s, error: %v, ", region, fetchErr)
			}

			mux.Lock()
			crossRegionResources = append(crossRegionResources, results)
			mux.Unlock()
		}(client, region)
	}

	wg.Wait()
	return crossRegionResources, err
}

func (w *multiRegionWrapper[T]) GetMultiRegionsClientMap() map[string]T {
	return w.clients
}

// GetRegions will initialize the singleton instance and perform the API request to retrieve the regions list only once, even if the function is called multiple times.
// Subsequent calls to the function will return the stored regions list without making another API request.
// In case of a failure the function returns the default region.
func getRegions(client DescribeCloudRegions, log *logp.Logger) []string {
	log.Debug("GetRegions starting...")

	once.Do(func() {
		log.Debug("Get aws regions for the first time")
		awsRegions = &cachedRegions{}

		output, err := client.DescribeRegions(context.Background(), nil)
		if err != nil {
			log.Errorf("failed DescribeRegions: %v", err)
			awsRegions.enabledRegions = []string{DefaultRegion}
			once = &sync.Once{} // reset singleton upon error
			return
		}

		for _, region := range output.Regions {
			awsRegions.enabledRegions = append(awsRegions.enabledRegions, *region.RegionName)
		}
	})

	log.Debugf("enabled regions for aws account, %v", awsRegions.enabledRegions)
	return awsRegions.enabledRegions
}
