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
	"errors"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/elb"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type ElbFetcher struct {
	log             *clog.Logger
	elbProvider     elb.LoadBalancerDescriber
	kubeClient      k8s.Interface
	lbRegexMatchers []*regexp.Regexp
	resourceCh      chan fetching.ResourceInfo
	cloudIdentity   *cloud.Identity
}

type ElbResource struct {
	lb       types.LoadBalancerDescription
	identity *cloud.Identity
}

func NewElbFetcher(log *clog.Logger, ch chan fetching.ResourceInfo, kubeProvider k8s.Interface, provider elb.LoadBalancerDescriber, identity *cloud.Identity, matchers string) *ElbFetcher {
	return &ElbFetcher{
		log:             log,
		elbProvider:     provider,
		cloudIdentity:   identity,
		kubeClient:      kubeProvider,
		lbRegexMatchers: []*regexp.Regexp{regexp.MustCompile(matchers)},
		resourceCh:      ch,
	}
}

func (f *ElbFetcher) Fetch(ctx context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Debug("Starting ElbFetcher.Fetch")

	balancers, err := f.GetLoadBalancers(ctx)
	if err != nil {
		return fmt.Errorf("failed to load balancers from Kubernetes %w", err)
	}
	result, err := f.elbProvider.DescribeLoadBalancers(ctx, balancers)
	if err != nil {
		return fmt.Errorf("failed to load balancers from ELB %w", err)
	}

	for _, loadBalancer := range result {
		f.resourceCh <- fetching.ResourceInfo{
			Resource:      ElbResource{lb: loadBalancer, identity: f.cloudIdentity},
			CycleMetadata: cycleMetadata,
		}
	}
	return nil
}

func (f *ElbFetcher) GetLoadBalancers(ctx context.Context) ([]string, error) {
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

func (f *ElbFetcher) Stop() {
}

func (r ElbResource) GetData() any {
	return r.lb
}

func (r ElbResource) GetMetadata() (fetching.ResourceMetadata, error) {
	if r.lb.LoadBalancerName == nil {
		return fetching.ResourceMetadata{}, errors.New("received nil pointer")
	}
	return fetching.ResourceMetadata{
		ID:      r.buildId(),
		Type:    fetching.CloudLoadBalancer,
		SubType: fetching.ElbType,
		Name:    *r.lb.LoadBalancerName,
	}, nil
}

// buildId A compromise because aws-sdk do not return an arn for an Elb
func (r ElbResource) buildId() string {
	id := fmt.Sprintf("%s-%s", r.identity.Account, *r.lb.LoadBalancerName)
	return id
}

func (r ElbResource) GetIds() []string {
	return []string{r.buildId()}
}

func (ElbResource) GetElasticCommonData() (map[string]any, error) {
	return map[string]any{
		"cloud.service.name": "ELB",
	}, nil
}
