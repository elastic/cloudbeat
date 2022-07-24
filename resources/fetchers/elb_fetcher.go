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
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/pkg/errors"
	"regexp"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	elbRegexTemplate = "([\\w-]+)-\\d+\\.%s.elb.amazonaws.com"
)

type ELBFetcher struct {
	log             *logp.Logger
	cfg             ELBFetcherConfig
	elbProvider     awslib.ElbLoadBalancerDescriber
	kubeClient      k8s.Interface
	lbRegexMatchers []*regexp.Regexp
	resourceCh      chan fetching.ResourceInfo
	cloudIdentity   *awslib.Identity
}

type ELBFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
	KubeConfig                    string `config:"Kubeconfig"`
}

type LoadBalancersDescription types.LoadBalancerDescription

type ELBResource struct {
	lb       LoadBalancersDescription
	identity *awslib.Identity
}

func (f *ELBFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting ELBFetcher.Fetch")

	balancers, err := f.GetLoadBalancers()
	if err != nil {
		return fmt.Errorf("failed to load balancers from Kubernetes %w", err)
	}
	result, err := f.elbProvider.DescribeLoadBalancer(ctx, balancers)
	if err != nil {
		return fmt.Errorf("failed to load balancers from ELB %w", err)
	}

	for _, loadBalancer := range result {
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      ELBResource{LoadBalancersDescription(loadBalancer), f.cloudIdentity},
			CycleMetadata: cMetadata,
		}
	}
	return nil
}

func (f *ELBFetcher) GetLoadBalancers() ([]string, error) {
	ctx := context.Background()
	services, err := f.kubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Kuberenetes services:  %w", err)
	}
	loadBalancers := make([]string, 0)
	for _, service := range services.Items {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			for _, matcher := range f.lbRegexMatchers {
				if matcher.MatchString(ingress.Hostname) {
					// Extract the repository name out of the image name
					lbName := matcher.FindStringSubmatch(ingress.Hostname)[1]
					loadBalancers = append(loadBalancers, lbName)
				}
			}
		}
	}
	return loadBalancers, nil
}

func (f *ELBFetcher) Stop() {
}

func (r ELBResource) GetData() interface{} {
	return r.lb
}

func (r ELBResource) GetMetadata() (fetching.ResourceMetadata, error) {
	if r.identity.Account == nil || r.lb.LoadBalancerName == nil {
		return fetching.ResourceMetadata{}, errors.New("received nil pointer")
	}

	return fetching.ResourceMetadata{
		// A compromise because aws-sdk do not return an arn for an ELB
		ID:      fmt.Sprintf("%s-%s", *r.identity.Account, *r.lb.LoadBalancerName),
		Type:    fetching.CloudLoadBalancer,
		SubType: fetching.ELBType,
		Name:    *r.lb.LoadBalancerName,
	}, nil
}
func (r ELBResource) GetElasticCommonData() any { return nil }
