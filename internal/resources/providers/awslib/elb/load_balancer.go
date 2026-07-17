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
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type ElasticLoadBalancerInfo struct {
	LoadBalancer types.LoadBalancerDescription `json:"load_balancer"`
	awsAccount   string
	region       string
	tags         map[string]string
	ipAddresses  []string
}

func (v ElasticLoadBalancerInfo) GetResourceArn() string {
	id := pointers.Deref(v.LoadBalancer.LoadBalancerName)
	if id == "" {
		return ""
	}
	return fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:loadbalancer/%s", v.region, v.awsAccount, id)
}

func (v ElasticLoadBalancerInfo) GetResourceName() string {
	return pointers.Deref(v.LoadBalancer.LoadBalancerName)
}

func (v ElasticLoadBalancerInfo) GetResourceType() string {
	return fetching.ElbType
}

func (v ElasticLoadBalancerInfo) GetRegion() string {
	return v.region
}

func (v ElasticLoadBalancerInfo) GetDNSName() string {
	return pointers.Deref(v.LoadBalancer.DNSName)
}

func (v ElasticLoadBalancerInfo) IsPubliclyAccessible() bool {
	return pointers.Deref(v.LoadBalancer.Scheme) == "internet-facing"
}

func (v ElasticLoadBalancerInfo) GetCreatedAt() *time.Time {
	return v.LoadBalancer.CreatedTime
}

// GetLoadBalancerType reports the load balancer type. Classic load balancers have no
// type field in the SDK, so we report a stable "classic" value.
func (v ElasticLoadBalancerInfo) GetLoadBalancerType() string {
	return "classic"
}

// GetState returns "active" for classic load balancers. The classic ELB API does not expose
// a state field, but any LB returned by DescribeLoadBalancers is live.
func (v ElasticLoadBalancerInfo) GetState() string {
	return "active"
}

// GetIPAddresses returns the IP addresses resolved from the load balancer's DNS name.
// Classic ELBs do not expose IPs directly via the AWS API; the addresses are resolved at
// fetch time (see Provider.DescribeAllLoadBalancers) and stored here.
func (v ElasticLoadBalancerInfo) GetIPAddresses() []string {
	return v.ipAddresses
}

// GetOwnerTag returns the value of the "Owner" tag (case-insensitive), if present.
func (v ElasticLoadBalancerInfo) GetOwnerTag() string {
	return awslib.LookupTag(v.tags, "owner")
}
