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

package logs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type Provider struct {
	clients map[string]Client
}

type Client interface {
	DescribeMetricFilters(ctx context.Context, params *cloudwatchlogs.DescribeMetricFiltersInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeMetricFiltersOutput, error)
}

func (p *Provider) DescribeMetricFilters(ctx context.Context, region *string, logGroup string) ([]types.MetricFilter, error) {
	var all []types.MetricFilter
	input := cloudwatchlogs.DescribeMetricFiltersInput{
		LogGroupName: aws.String(logGroup),
	}
	client, err := awslib.GetClient(region, p.clients)
	if err != nil {
		return nil, err
	}
	for {
		output, err := client.DescribeMetricFilters(ctx, &input)
		if err != nil {
			return nil, err
		}
		all = append(all, output.MetricFilters...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return all, nil
}
