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

package cloudwatch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type Provider struct {
	clients map[string]Client
}

type Client interface {
	DescribeAlarms(ctx context.Context, params *cloudwatch.DescribeAlarmsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.DescribeAlarmsOutput, error)
}

func (p *Provider) DescribeAlarms(ctx context.Context, region *string, filters []string) ([]types.MetricAlarm, error) {
	var all []types.MetricAlarm
	input := cloudwatch.DescribeAlarmsInput{}
	client, err := awslib.GetClient(region, p.clients)
	if err != nil {
		return nil, err
	}
	for {
		output, err := client.DescribeAlarms(ctx, &input)
		if err != nil {
			return nil, err
		}
		for _, metric := range output.MetricAlarms {
			for _, filter := range filters {
				if metric.MetricName != nil && filter == *metric.MetricName {
					all = append(all, metric)
					break
				}
			}
		}
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}
	return all, nil
}
