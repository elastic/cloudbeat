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

package fetchers

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
)

type ELBProvider struct {
	client *elasticloadbalancing.Client
}

func NewELBProvider(cfg aws.Config) *ELBProvider {
	svc := elasticloadbalancing.New(cfg)
	return &ELBProvider{
		client: svc,
	}
}

// DescribeLoadBalancer method will return up to 400 results
// If we will ever want to increase this number, DescribeLoadBalancers support paginated requests
func (provider ELBProvider) DescribeLoadBalancer(ctx context.Context, balancersNames []string) ([]elasticloadbalancing.LoadBalancerDescription, error) {
	logp.Info("elb fetcher started")

	input := &elasticloadbalancing.DescribeLoadBalancersInput{
		LoadBalancerNames: balancersNames,
	}

	req := provider.client.DescribeLoadBalancersRequest(input)
	response, err := req.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to describe load balancers %s from elb, error - %w", balancersNames, err)
	}

	return response.LoadBalancerDescriptions, err
}
