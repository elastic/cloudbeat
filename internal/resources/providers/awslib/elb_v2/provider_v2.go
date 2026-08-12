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

package elb_v2

import (
	"context"
	"fmt"

	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func (p *Provider) DescribeLoadBalancers(ctx context.Context) ([]awslib.AwsResource, error) {
	p.log.Debug("Fetching Elastic Load Balancers")
	elbs, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &elbv2.DescribeLoadBalancersInput{}
		all := []types.LoadBalancer{}
		for {
			output, err := c.DescribeLoadBalancers(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.LoadBalancers...)
			if output.NextMarker == nil {
				break
			}
			input.Marker = output.NextMarker
		}

		lbs := make([]*ElasticLoadBalancerInfo, 0, len(all))
		arns := make([]string, 0, len(all))
		for _, item := range all {
			loadBalancer := &ElasticLoadBalancerInfo{
				LoadBalancer: item,
				region:       region,
			}
			listeners, err := p.describeListeners(ctx, region, loadBalancer.GetResourceArn())
			if err != nil {
				p.log.Errorf("Error fetching listeners for %s: %v", loadBalancer.GetResourceArn(), err)
			} else {
				loadBalancer.Listeners = listeners
			}
			lbs = append(lbs, loadBalancer)
			if arn := loadBalancer.GetResourceArn(); arn != "" {
				arns = append(arns, arn)
			}
		}

		tagsByArn := p.describeTags(ctx, c, arns)

		result := make([]awslib.AwsResource, 0, len(lbs))
		for _, lb := range lbs {
			lb.tags = tagsByArn[lb.GetResourceArn()]
			result = append(result, lb)
		}
		return result, nil
	})
	result := lo.Flatten(elbs)
	if err != nil {
		p.log.Debugf("Fetched %d Elastic Load Balancers", len(result))
	}
	return result, err
}

// describeTags fetches tags for the given load balancer ARNs (chunked to the AWS 20-ARN
// limit) and returns a map of load balancer ARN to its tag key/value pairs.
func (p *Provider) describeTags(ctx context.Context, c Client, arns []string) map[string]map[string]string {
	out := map[string]map[string]string{}
	for _, chunk := range lo.Chunk(arns, 20) {
		if len(chunk) == 0 {
			continue
		}
		resp, err := c.DescribeTags(ctx, &elbv2.DescribeTagsInput{ResourceArns: chunk})
		if err != nil {
			p.log.Errorf("Could not fetch tags for load balancers: %v", err)
			continue
		}
		for _, td := range resp.TagDescriptions {
			arn := pointers.Deref(td.ResourceArn)
			if arn == "" {
				continue
			}
			tags := make(map[string]string, len(td.Tags))
			for _, t := range td.Tags {
				tags[pointers.Deref(t.Key)] = pointers.Deref(t.Value)
			}
			out[arn] = tags
		}
	}
	return out
}

// describeListeners queries and returns all Listeners filtered by ELB ARN and region.
// Used by DescribeLoadBalancers to find Listeners connected to a specific Elastic Load
// Balancer (v2).
func (p *Provider) describeListeners(ctx context.Context, region, loadBalancerArn string) ([]types.Listener, error) {
	p.log.Debugf("Fetching ELB Listeners for %s", loadBalancerArn)
	c, ok := p.clients[region]
	if !ok {
		return nil, fmt.Errorf("could not find client for %s region", region)
	}
	input := &elbv2.DescribeListenersInput{
		LoadBalancerArn: pointers.Ref(loadBalancerArn),
	}
	var result []types.Listener
	for {
		output, err := c.DescribeListeners(ctx, input)
		if err != nil {
			return nil, err
		}
		result = append(result, output.Listeners...)
		if output.NextMarker == nil {
			break
		}
		input.Marker = output.NextMarker
	}
	p.log.Debugf("Fetched %d ELB Listeners", len(result))
	return result, nil
}
