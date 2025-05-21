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
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/elb"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

const (
	elbRegex = "([\\w-]+)-\\d+\\.us-east-2.elb.amazonaws.com"
)

type ElbFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestElbFetcherTestSuite(t *testing.T) {
	s := new(ElbFetcherTestSuite)

	suite.Run(t, s)
}

func (s *ElbFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *ElbFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *ElbFetcherTestSuite) TestCreateFetcher() {
	var (
		testAccount = "test-account"
		lbName      = "adda9cdc89b13452e92d48be46858d37"
	)

	var tests = []struct {
		ns                  string
		loadBalancerIngress []v1.LoadBalancerIngress
		lbResponse          []types.LoadBalancerDescription
		expectedlbNames     []string
	}{
		{
			"test_namespace",
			[]v1.LoadBalancerIngress{
				{
					Hostname: "adda9cdc89b13452e92d48be46858d37-1423035038.us-east-2.elb.amazonaws.com",
				},
			},
			[]types.LoadBalancerDescription{{
				Instances:        []types.Instance{},
				LoadBalancerName: &lbName,
			}},
			[]string{lbName},
		},
		{
			"test_namespace",
			[]v1.LoadBalancerIngress{
				{
					Hostname: "adda9cdc89b13452e92d48be46858d37-1423035038.wrong-region.elb.amazonaws.com",
				},
			},
			[]types.LoadBalancerDescription{},
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
		t := s.T()
		_, err := kubeclient.CoreV1().Services(test.ns).Create(t.Context(), services, metav1.CreateOptions{})
		s.Require().NoError(err)

		elbProvider := &elb.MockLoadBalancerDescriber{}
		elbProvider.EXPECT().DescribeLoadBalancers(mock.Anything, mock.Anything).Return(test.lbResponse, nil)

		identity := cloud.Identity{
			Account: testAccount,
		}

		regexMatchers := []*regexp.Regexp{regexp.MustCompile(elbRegex)}

		elbFetcher := ElbFetcher{
			log:             testhelper.NewLogger(s.T()),
			elbProvider:     elbProvider,
			kubeClient:      kubeclient,
			lbRegexMatchers: regexMatchers,
			resourceCh:      s.resourceCh,
			cloudIdentity:   &identity,
		}

		err = elbFetcher.Fetch(t.Context(), cycle.Metadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Equal(len(test.expectedlbNames), len(results))
		s.Require().NoError(err)

		for i, expectedLbName := range test.expectedlbNames {
			elbResource := results[i].Resource.(ElbResource)
			metadata, err := elbResource.GetMetadata()

			s.Require().NoError(err)
			s.Equal(expectedLbName, *elbResource.lb.LoadBalancerName)
			s.Equal(*elbResource.lb.LoadBalancerName, metadata.Name)
			s.Equal(fmt.Sprintf("%s-%s", testAccount, *elbResource.lb.LoadBalancerName), metadata.ID)
		}
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
		t := s.T()
		_, err := kubeclient.CoreV1().Services(test.ns).Create(t.Context(), services, metav1.CreateOptions{})
		s.Require().NoError(err)

		elbProvider := &elb.MockLoadBalancerDescriber{}
		elbProvider.EXPECT().DescribeLoadBalancers(mock.Anything, mock.Anything).Return(nil, test.error)

		regexMatchers := []*regexp.Regexp{regexp.MustCompile(elbRegex)}

		elbFetcher := ElbFetcher{
			log:             testhelper.NewLogger(s.T()),
			elbProvider:     elbProvider,
			kubeClient:      kubeclient,
			lbRegexMatchers: regexMatchers,
			resourceCh:      s.resourceCh,
			cloudIdentity:   nil,
		}

		ctx := t.Context()

		err = elbFetcher.Fetch(ctx, cycle.Metadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Nil(results)
		s.Require().EqualError(err, fmt.Sprintf("failed to load balancers from ELB %s", test.error.Error()))
	}
}

func (s *ElbFetcherTestSuite) TestElbResource_GetMetadata() {
	r := ElbResource{
		identity: &cloud.Identity{
			Account: "test-account",
		},
		lb: types.LoadBalancerDescription{
			LoadBalancerName: aws.String("test-lb-name"),
		},
	}
	meta, err := r.GetMetadata()
	s.Require().NoError(err)
	s.Equal(fetching.ResourceMetadata{ID: "test-account-test-lb-name", Type: "load-balancer", SubType: "aws-elb", Name: "test-lb-name"}, meta)
	m, err := r.GetElasticCommonData()
	s.Require().NoError(err)
	s.Len(m, 1)
	s.Contains(m, "cloud.service.name")
}
