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
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elastic/elastic-agent-libs/logp"
	"sync"
)

var (
	instance *singleton
	once     = &sync.Once{}
)

type AWSCommonUtil interface {
	DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)
}

type MultiRegionWrapper[T any] struct {
	Clients map[string]T
}

// Fetch retrieves resources from multiple regions concurrently using the provided fetcher function.
func (w *MultiRegionWrapper[T]) Fetch(fetcher func(T) ([]AwsResource, error)) ([]AwsResource, error) {
	var err error
	var wg sync.WaitGroup
	var mux sync.Mutex
	var crossRegionResources []AwsResource

	for region, client := range w.Clients {
		wg.Add(1)
		go func(client T, region string) {
			defer wg.Done()
			results, fetchErr := fetcher(client)
			if fetchErr != nil {
				err = fmt.Errorf("Fail to retrieve aws resources for region: %s, error: %v, ", region, fetchErr)
			}

			mux.Lock()
			crossRegionResources = append(crossRegionResources, results...)
			mux.Unlock()
		}(client, region)
	}

	wg.Wait()
	return crossRegionResources, err
}

// CreateMultiRegionClients is a utility function that is used to create a map of client instances of a given type T for multiple regions.
func CreateMultiRegionClients[T any](client AWSCommonUtil, cfg awssdk.Config, factory func(cfg awssdk.Config) T, log *logp.Logger) *MultiRegionWrapper[T] {
	var clientsMap = make(map[string]T, 0)
	for _, region := range getRegions(client, log) {
		cfg.Region = region
		clientsMap[region] = factory(cfg)
	}

	w := &MultiRegionWrapper[T]{
		Clients: clientsMap,
	}

	return w
}

// GetRegions will initialize the singleton instance and perform the API request to retrieve the regions list only once, even if the function is called multiple times.
// Subsequent calls to the function will return the stored regions list without making another API request.
// In case of a failure the function returns the default region.
func getRegions(client AWSCommonUtil, log *logp.Logger) []string {
	log.Debug("GetRegions starting...")
	var mu sync.Mutex
	var initErr error

	mu.Lock()
	defer mu.Unlock()
	once.Do(func() {
		log.Debug("Get aws regions for the first time")
		instance = &singleton{}

		output, err := client.DescribeRegions(context.Background(), nil)
		if err != nil {
			initErr = fmt.Errorf("failed DescribeRegions: %w", err)
			once = &sync.Once{} // reset singleton upon error
			return
		}

		for _, region := range output.Regions {
			instance.regions = append(instance.regions, *region.RegionName)
		}
	})

	if initErr != nil {
		log.Errorf("Error: %v, init client only for the default region: %s", initErr, DefaultRegion)
		instance.regions = append(instance.regions, DefaultRegion)
	}

	log.Debugf("enabled regions for aws account, %v", instance.regions)
	return instance.regions
}
