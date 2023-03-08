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
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/mapstr"

	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
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
	var (
		testAccount = "test-account"
		testID      = "test-id"
		testARN     = "test-arn"
		lbName      = "adda9cdc89b13452e92d48be46858d37"
	)

	tests := []struct {
		ns                  string
		loadBalancerIngress []v1.LoadBalancerIngress
		lbResponse          awslib.ElbLoadBalancerDescriptions
		expectedlbNames     []string
	}{
		{
			"test_namespace",
			[]v1.LoadBalancerIngress{
				{
					Hostname: "adda9cdc89b13452e92d48be46858d37-1423035038.us-east-2.elb.amazonaws.com",
				},
			},
			awslib.ElbLoadBalancerDescriptions{{
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
			awslib.ElbLoadBalancerDescriptions{},
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
		s.NoError(err)

		mockedKubernetesClientGetter := &providers.MockKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything, mock.Anything).Return(kubeclient, nil)

		elbProvider := &awslib.MockElbLoadBalancerDescriber{}
		elbProvider.EXPECT().DescribeLoadBalancer(mock.Anything, mock.Anything).Return(test.lbResponse, nil)

		identity := awslib.Identity{
			Account: &testAccount,
			Arn:     &testARN,
			UserId:  &testID,
		}

		elbFetcher := New(
			WithLogger(s.log),
			WithConfig(&config.Config{
				Fetchers: []*agentconfig.C{
					agentconfig.MustNewConfigFrom(mapstr.M{
						"name": "aws-elb",
					}),
				},
			}),
			WithElbProvider(elbProvider),
			WithKubeClient(kubeclient),
			WithRegexMatcher(""),
			WithResourceChan(s.resourceCh),
			WithCloudIdentity(&identity),
		)

		err = elbFetcher.Fetch(context.Background(), fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Equal(len(test.expectedlbNames), len(results))
		s.NoError(err)

		for i, expectedLbName := range test.expectedlbNames {
			elbResource := results[i].Resource.(ElbResource)
			metadata, err := elbResource.GetMetadata()

			s.NoError(err)
			s.Equal(expectedLbName, *elbResource.lb.LoadBalancerName)
			s.Equal(*elbResource.lb.LoadBalancerName, metadata.Name)
			s.Equal(fmt.Sprintf("%s-%s", testAccount, *elbResource.lb.LoadBalancerName), metadata.ID)
		}
	}
}

func (s *ElbFetcherTestSuite) TestCreateFetcherErrorCases() {
	tests := []struct {
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
			fmt.Errorf("elb error"),
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
		s.NoError(err)

		mockedKubernetesClientGetter := &providers.MockKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything, mock.Anything).Return(kubeclient, nil)

		elbProvider := &awslib.MockElbLoadBalancerDescriber{}
		elbProvider.EXPECT().DescribeLoadBalancer(mock.Anything, mock.Anything).Return(nil, test.error)

		elbFetcher := New(
			WithLogger(s.log),
			WithConfig(&config.Config{
				Fetchers: []*agentconfig.C{
					agentconfig.MustNewConfigFrom(mapstr.M{
						"name": "aws-elb",
					}),
				},
			}),
			WithElbProvider(elbProvider),
			WithKubeClient(kubeclient),
			WithRegexMatcher(""),
			WithResourceChan(s.resourceCh),
		)

		ctx := context.Background()

		err = elbFetcher.Fetch(ctx, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Nil(results)
		s.EqualError(err, fmt.Sprintf("failed to load balancers from ELB %s", test.error.Error()))
	}
}
