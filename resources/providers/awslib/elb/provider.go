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

package elb

import (
	"context"
	"fmt"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
)

// DescribeLoadBalancer returns LoadBalancerDescriptions which contain information about the load balancers.
// When balancersNames is empty, it will describe all the existing load balancers
func (provider Provider) DescribeLoadBalancer(ctx context.Context, balancersNames []string) ([]types.LoadBalancerDescription, error) {
	input := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: balancersNames,
	}

	var loadBalancerDescriptions []types.LoadBalancerDescription
	paginator := elb.NewDescribeLoadBalancersPaginator(provider.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error DescribeRepositories with Paginator: %w", err)
		}
		loadBalancerDescriptions = append(loadBalancerDescriptions, page.LoadBalancerDescriptions...)
	}

	return loadBalancerDescriptions, nil
}
