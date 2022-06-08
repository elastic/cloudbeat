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
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

type ElbFactoryTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestElbFactoryTestSuite(t *testing.T) {
	s := new(ElbFactoryTestSuite)
	s.log = logp.NewLogger("cloudbeat_elb_factory_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ElbFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config        string
		region        string
		expectedRegex string
	}{
		{
			`
name: aws-elb
`,
			"us1-east",
			"([\\w-]+)-\\d+\\.us1-east.elb.amazonaws.com",
		},
	}

	for _, test := range tests {

		kubeclient := k8sfake.NewSimpleClientset()
		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		awsConfig := awslib.Config{Config: aws.Config{
			Region: test.region,
		}}

		elbProvider := &awslib.MockedELBLoadBalancerDescriber{}

		factory := &ELBFactory{
			extraElements: func() (elbExtraElements, error) {
				return elbExtraElements{
					balancerDescriber:      elbProvider,
					awsConfig:              awsConfig,
					kubernetesClientGetter: mockedKubernetesClientGetter,
				}, nil
			},
		}

		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(s.log, cfg, nil)
		s.NoError(err)
		s.NotNil(fetcher)

		elbFetcher, ok := fetcher.(*ELBFetcher)
		s.True(ok)
		s.Equal(elbProvider, elbFetcher.elbProvider)
		s.Equal(kubeclient, elbFetcher.kubeClient)
		s.Equal(test.expectedRegex, elbFetcher.lbRegexMatchers[0].String())
	}
}
