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
	ec2imds "github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/elastic/elastic-agent-libs/logp"
)

var currentSingleSelector = newCurrentRegionSelector()

type currentRegionSelector struct {
	once   *sync.Once
	lock   *sync.Mutex
	cache  *cachedRegions
	client currentCloudRegion
}

type currentCloudRegion interface {
	GetMetadata(ctx context.Context, cfg aws.Config) (*ec2imds.InstanceIdentityDocument, error)
}

func CurrentRegionSelector() *currentRegionSelector {
	return currentSingleSelector
}

func newCurrentRegionSelector() *currentRegionSelector {
	return &currentRegionSelector{
		once:  &sync.Once{},
		lock:  &sync.Mutex{},
		cache: &cachedRegions{},
	}
}

func (s *currentRegionSelector) Regions(ctx context.Context, cfg aws.Config) ([]string, error) {
	log := logp.NewLogger("aws")
	log.Debug("currentRegionSelector starting...")
	var err error

	// Make sure that consequent calls to the function will keep trying to retrieve the region until it succeeds.
	s.lock.Lock()
	defer s.lock.Unlock()
	s.once.Do(func() {
		if s.client == nil {
			s.client = &Ec2MetadataProvider{}
		}

		var metadata *Ec2Metadata
		metadata, err = s.client.GetMetadata(ctx, cfg)
		if err != nil {
			log.Errorf("failed DescribeRegions: %v", err)
			s.once = &sync.Once{} // reset singleton upon error
			return
		}

		s.cache.regions = []string{metadata.Region}
	})

	log.Debugf("current region of aws instance, %v", s.cache.regions)
	return s.cache.regions, err
}
