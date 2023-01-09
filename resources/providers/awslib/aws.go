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

const DefaultRegion = "us-east-1"

var (
	instance *singleton
	once     = &sync.Once{}
)

type singleton struct {
	regions []string
}

type Config struct {
	Config awssdk.Config
}

type AwsResource interface {
	GetResourceArn() string
	GetResourceName() string
	GetResourceType() string
}

type AWSCommonUtil interface {
	DescribeRegions(ctx context.Context, params *ec2.DescribeRegionsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRegionsOutput, error)
}

// GetRegions will initialize the singleton instance and perform the API request to retrieve the regions list only once, even if the function is called multiple times.
// Subsequent calls to the function will return the stored regions list without making another API request.
func GetRegions(log *logp.Logger, client AWSCommonUtil) ([]string, error) {
	log.Debug("GetRegions starting...")

	var initErr error
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
	return instance.regions, initErr
}
