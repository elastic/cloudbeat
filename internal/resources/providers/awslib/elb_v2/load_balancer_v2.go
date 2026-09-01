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
	"time"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type ElasticLoadBalancerInfo struct {
	LoadBalancer types.LoadBalancer `json:"load_balancer"`
	Listeners    []types.Listener   `json:"listeners"`
	region       string
	tags         map[string]string
	dnsResolvedIPs []string
}

// NewElasticLoadBalancerInfo constructs an ELBv2 wrapper. dnsResolvedIPs are the addresses
// resolved from the load balancer's DNS name when the AWS API returns none (see Provider).
func NewElasticLoadBalancerInfo(lb types.LoadBalancer, region string, tags map[string]string, dnsResolvedIPs []string) *ElasticLoadBalancerInfo {
	return &ElasticLoadBalancerInfo{
		LoadBalancer:   lb,
		region:         region,
		tags:           tags,
		dnsResolvedIPs: dnsResolvedIPs,
	}
}

func (v ElasticLoadBalancerInfo) GetResourceArn() string {
	return pointers.Deref(v.LoadBalancer.LoadBalancerArn)
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
	return v.LoadBalancer.Scheme == types.LoadBalancerSchemeEnumInternetFacing
}

func (v ElasticLoadBalancerInfo) GetCreatedAt() *time.Time {
	return v.LoadBalancer.CreatedTime
}

// GetLoadBalancerType reports the load balancer type (application, network, gateway).
func (v ElasticLoadBalancerInfo) GetLoadBalancerType() string {
	return string(v.LoadBalancer.Type)
}

// GetState reports the load balancer state code (e.g. active, provisioning).
func (v ElasticLoadBalancerInfo) GetState() string {
	if v.LoadBalancer.State == nil {
		return ""
	}
	return string(v.LoadBalancer.State.Code)
}

// GetIPAddresses returns the IP addresses of the load balancer. NLBs with Elastic IPs
// expose them via the AWS API (IpAddress, PrivateIPv4Address, IPv6Address). When the API
// returns nothing — which is the common case for internet-facing ALBs and most NLBs — the
// provider resolves the DNSName at fetch time and stores the result in dnsResolvedIPs.
func (v ElasticLoadBalancerInfo) GetIPAddresses() []string {
	var ips []string
	for _, az := range v.LoadBalancer.AvailabilityZones {
		for _, addr := range az.LoadBalancerAddresses {
			if ip := pointers.Deref(addr.IpAddress); ip != "" {
				ips = append(ips, ip)
			}
			if ip := pointers.Deref(addr.PrivateIPv4Address); ip != "" {
				ips = append(ips, ip)
			}
			if ip := pointers.Deref(addr.IPv6Address); ip != "" {
				ips = append(ips, ip)
			}
		}
	}
	if len(ips) == 0 {
		return v.dnsResolvedIPs
	}
	return ips
}

// GetOwnerTag returns the value of the "Owner" tag (case-insensitive), if present.
func (v ElasticLoadBalancerInfo) GetOwnerTag() string {
	return awslib.LookupTag(v.tags, "owner")
}
