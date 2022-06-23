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
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

const (
	elbRegex = "([\\w-]+)-\\d+\\.us-east-2.elb.amazonaws.com"
)

type ElbFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

func TestElbFetcherTestSuite(t *testing.T) {
	s := new(ElbFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_elb_fetcher_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ElbFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *ElbFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *ElbFetcherTestSuite) TestCreateFetcher() {

	var tests = []struct {
		ns                  string
		loadBalancerIngress []v1.LoadBalancerIngress
		lbResponse          []elasticloadbalancing.LoadBalancerDescription
		expectedlbNames     []string
	}{
		{
			"test_namespace",
			[]v1.LoadBalancerIngress{
				{
					Hostname: "adda9cdc89b13452e92d48be46858d37-1423035038.us-east-2.elb.amazonaws.com",
				},
			},
			[]elasticloadbalancing.LoadBalancerDescription{{
				Instances: []elasticloadbalancing.Instance{},
			}},
			[]string{"adda9cdc89b13452e92d48be46858d37"},
		},
		{
			"test_namespace",
			[]v1.LoadBalancerIngress{
				{
					Hostname: "adda9cdc89b13452e92d48be46858d37-1423035038.wrong-region.elb.amazonaws.com",
				},
			},
			[]elasticloadbalancing.LoadBalancerDescription{},
			[]string{},
		},
	}
	for _, test := range tests {
		kubeclient := k8sfake.NewSimpleClientset()

		services := &v1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "apps/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testing_pod",
				Namespace: test.ns,
			},
			Status: v1.ServiceStatus{
				LoadBalancer: v1.LoadBalancerStatus{
					Ingress: test.loadBalancerIngress,
				},
			},
			Spec: v1.ServiceSpec{},
		}
		_, err := kubeclient.CoreV1().Services(test.ns).Create(context.Background(), services, metav1.CreateOptions{})
		s.Nil(err)

		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		elbProvider := &awslib.MockedELBLoadBalancerDescriber{}
		elbProvider.EXPECT().DescribeLoadBalancer(mock.Anything, mock.MatchedBy(func(balancers []string) bool {
			return s.Equal(balancers, test.expectedlbNames)
		})).Return(test.lbResponse, nil)

		regexMatchers := []*regexp.Regexp{regexp.MustCompile(elbRegex)}

		elbFetcher := ELBFetcher{
			log:             s.log,
			cfg:             ELBFetcherConfig{},
			elbProvider:     elbProvider,
			kubeClient:      kubeclient,
			lbRegexMatchers: regexMatchers,
			resourceCh:      s.resourceCh,
		}

		ctx := context.Background()

		expectedResource := ELBResource{test.lbResponse}
		err = elbFetcher.Fetch(ctx, fetching.CycleMetadata{})

		results := testhelper.CollectResources(s.resourceCh)
		elbResource := results[0].Resource.(ELBResource)

		s.Equal(1, len(results))
		s.Equal(expectedResource, elbResource)
		s.Nil(err)
	}
}

func (s *ElbFetcherTestSuite) TestCreateFetcherErrorCases() {

	var tests = []struct {
		ns                  string
		loadBalancerIngress []v1.LoadBalancerIngress
		error               error
	}{
		{
			"test_namespace",
			[]v1.LoadBalancerIngress{
				{
					Hostname: "adda9cdc89b13452e92d48be46858d37-1423035038.us-east-2.elb.amazonaws.com",
				},
			},
			fmt.Errorf("elb error")},
	}
	for _, test := range tests {
		kubeclient := k8sfake.NewSimpleClientset()

		services := &v1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "apps/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testing_pod",
				Namespace: test.ns,
			},
			Status: v1.ServiceStatus{
				LoadBalancer: v1.LoadBalancerStatus{
					Ingress: test.loadBalancerIngress,
				},
			},
			Spec: v1.ServiceSpec{},
		}
		_, err := kubeclient.CoreV1().Services(test.ns).Create(context.Background(), services, metav1.CreateOptions{})
		s.Nil(err)

		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		elbProvider := &awslib.MockedELBLoadBalancerDescriber{}
		elbProvider.EXPECT().DescribeLoadBalancer(mock.Anything, mock.Anything).Return(nil, test.error)

		regexMatchers := []*regexp.Regexp{regexp.MustCompile(elbRegex)}

		elbFetcher := ELBFetcher{
			log:             s.log,
			cfg:             ELBFetcherConfig{},
			elbProvider:     elbProvider,
			kubeClient:      kubeclient,
			lbRegexMatchers: regexMatchers,
			resourceCh:      s.resourceCh,
		}

		ctx := context.Background()

		err = elbFetcher.Fetch(ctx, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Nil(results)
		s.EqualError(err, fmt.Sprintf("failed to load balancers from ELB %s", test.error.Error()))
	}
}
