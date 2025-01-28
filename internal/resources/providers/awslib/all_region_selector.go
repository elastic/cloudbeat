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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

type describeCloudRegions interface {
	DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)
}

type allRegionSelector struct {
	client describeCloudRegions
}

// Regions will initialize the singleton instance and perform the API request to retrieve the regions list only once, even if the function is called multiple times.
// Subsequent calls to the function will return the stored regions list without making another API request.
// In case of a failure the function returns an error and resets the singleton instance.
func (s *allRegionSelector) Regions(ctx context.Context, cfg aws.Config) ([]string, error) {
	log := clog.NewLogger("aws")
	log.Info("Getting all available regions for the current account")

	if s.client == nil {
		s.client = ec2.NewFromConfig(cfg)
	}

	output, err := s.client.DescribeRegions(ctx, nil)
	if err != nil {
		log.Errorf("Failed getting available regions: %v", err)
		return nil, err
	}

	result := []string{}
	for _, region := range output.Regions {
		result = append(result, *region.RegionName)
	}

	log.Infof("Available regions, %+q", result)
	return result, nil
}
