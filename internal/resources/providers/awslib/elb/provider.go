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
	"sort"

	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

// DescribeLoadBalancers returns LoadBalancerDescriptions which contain information about the load balancers.
// When balancersNames is empty, it will describe all the existing load balancers
func (p *Provider) DescribeLoadBalancers(ctx context.Context, balancersNames []string) ([]types.LoadBalancerDescription, error) {
	p.log.Debug("Fetching Classic Elastic Load Balancers")
	input := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: balancersNames,
	}

	var loadBalancerDescriptions []types.LoadBalancerDescription
	paginator := elb.NewDescribeLoadBalancersPaginator(p.client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error DescribeRepositories with Paginator: %w", err)
		}
		loadBalancerDescriptions = append(loadBalancerDescriptions, page.LoadBalancerDescriptions...)
	}

	p.log.Debugf("Fetched %d Classic Elastic Load Balancers", len(loadBalancerDescriptions))
	return loadBalancerDescriptions, nil
}

func (p *Provider) DescribeAllLoadBalancers(ctx context.Context) ([]awslib.AwsResource, error) {
	p.log.Debug("Fetching Classic Elastic Load Balancers")
	elbs, err := awslib.MultiRegionFetch(ctx, p.clients, func(ctx context.Context, region string, c Client) ([]awslib.AwsResource, error) {
		input := &elb.DescribeLoadBalancersInput{}
		all := []types.LoadBalancerDescription{}
		for {
			output, err := c.DescribeLoadBalancers(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.LoadBalancerDescriptions...)
			if output.NextMarker == nil {
				break
			}
			input.Marker = output.NextMarker
		}

		names := make([]string, 0, len(all))
		for _, item := range all {
			if n := pointers.Deref(item.LoadBalancerName); n != "" {
				names = append(names, n)
			}
		}
		tagsByName := p.describeTags(ctx, c, names)

		var result []awslib.AwsResource
		for _, item := range all {
			info := &ElasticLoadBalancerInfo{
				LoadBalancer: item,
				awsAccount:   p.awsAccountID,
				region:       region,
				tags:         tagsByName[pointers.Deref(item.LoadBalancerName)],
			}
			if dnsName := pointers.Deref(item.DNSName); dnsName != "" {
				if ips, err := p.resolver.LookupHost(ctx, dnsName); err != nil {
					p.log.Debugf("Could not resolve IPs for classic ELB %q: %v", dnsName, err)
				} else {
					sort.Strings(ips)
					info.ipAddresses = ips
				}
			}
			result = append(result, info)
		}
		return result, nil
	})
	result := lo.Flatten(elbs)
	if err != nil {
		p.log.Debugf("Fetched %d Classic Elastic Load Balancers", len(result))
	}
	return lo.Flatten(elbs), err
}

// describeTags fetches tags for the given classic load balancer names (chunked to the AWS
// 20-name limit) and returns a map of load balancer name to its tag key/value pairs.
func (p *Provider) describeTags(ctx context.Context, c Client, names []string) map[string]map[string]string {
	out := map[string]map[string]string{}
	for _, chunk := range lo.Chunk(names, 20) {
		if len(chunk) == 0 {
			continue
		}
		resp, err := c.DescribeTags(ctx, &elb.DescribeTagsInput{LoadBalancerNames: chunk})
		if err != nil {
			p.log.Errorf("Could not fetch tags for classic load balancers: %v", err)
			continue
		}
		for _, td := range resp.TagDescriptions {
			name := pointers.Deref(td.LoadBalancerName)
			if name == "" {
				continue
			}
			tags := make(map[string]string, len(td.Tags))
			for _, t := range td.Tags {
				tags[pointers.Deref(t.Key)] = pointers.Deref(t.Value)
			}
			out[name] = tags
		}
	}
	return out
}
