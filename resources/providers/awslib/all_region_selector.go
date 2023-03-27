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
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elastic/elastic-agent-libs/logp"
)

var allSingleSelector = newAllRegionSelector()

type cachedRegions struct {
	regions []string
}

type describeCloudRegions interface {
	DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)
}

func AllRegionSelector() *allRegionsSelector {
	return allSingleSelector
}

func newAllRegionSelector() *allRegionsSelector {
	return &allRegionsSelector{
		once:  &sync.Once{},
		lock:  &sync.Mutex{},
		cache: &cachedRegions{},
	}
}

type allRegionsSelector struct {
	once   *sync.Once
	lock   *sync.Mutex
	cache  *cachedRegions
	client describeCloudRegions
}

// Regions will initialize the singleton instance and perform the API request to retrieve the regions list only once, even if the function is called multiple times.
// Subsequent calls to the function will return the stored regions list without making another API request.
// In case of a failure the function returns an error and resets the singleton instance.
func (s *allRegionsSelector) Regions(ctx context.Context, cfg aws.Config) ([]string, error) {
	log := logp.NewLogger("aws")
	log.Debug("allRegionsSelector starting...")
	var err error

	// Make sure that consequent calls to the function will keep trying to retrieve the regions list until it succeeds.
	s.lock.Lock()
	defer s.lock.Unlock()
	s.once.Do(func() {
		if s.client == nil {
			s.client = ec2.NewFromConfig(cfg)
		}

		log.Debug("Get aws regions for the first time")
		var output *ec2.DescribeRegionsOutput
		output, err = s.client.DescribeRegions(ctx, nil)
		if err != nil {
			log.Errorf("failed DescribeRegions: %v", err)
			s.once = &sync.Once{} // reset singleton upon error
			return
		}

		s.cache = &cachedRegions{}
		for _, region := range output.Regions {
			s.cache.regions = append(s.cache.regions, *region.RegionName)
		}
	})

	log.Debugf("enabled regions for aws account, %v", s.cache.regions)
	return s.cache.regions, err
}
