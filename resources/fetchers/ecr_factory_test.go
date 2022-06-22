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

type EcrFactoryTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestEcrFactoryTestSuite(t *testing.T) {
	s := new(EcrFactoryTestSuite)
	s.log = logp.NewLogger("cloudbeat_ecr_factory_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *EcrFactoryTestSuite) TestCreateFetcher() {
	var tests = []struct {
		config        string
		region        string
		account       string
		expectedRegex []string
	}{
		{
			`
name: aws-ecr
`,
			"us1-east",
			"my-account",
			[]string{
				"^my-account\\.dkr\\.ecr\\.us1-east\\.amazonaws\\.com\\/([-\\w\\.\\/]+)[:,@]?",
				"public\\.ecr\\.aws\\/\\w+\\/([-\\w\\.\\/]+)\\:?",
			},
		},
	}

	for _, test := range tests {

		kubeclient := k8sfake.NewSimpleClientset()
		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		identity := awslib.Identity{
			Account: &test.account,
		}
		identityProvider := &awslib.MockedIdentityProviderGetter{}
		identityProvider.EXPECT().GetIdentity(mock.Anything).Return(&identity, nil)
		awsConfig := awslib.Config{Config: aws.Config{
			Region: test.region,
		}}
		awsconfigProvider := &awslib.MockConfigGetter{}
		awsconfigProvider.EXPECT().GetConfig().Return(awsConfig)

		ecrProvider := &awslib.MockedEcrRepositoryDescriber{}
		ecrPublicProvider := &awslib.MockedEcrRepositoryDescriber{}

		factory := &ECRFactory{
			extraElements: func() (ecrExtraElements, error) {
				return ecrExtraElements{
					awsConfig:               awsConfig,
					kubernetesClientGetter:  mockedKubernetesClientGetter,
					identityProviderGetter:  identityProvider,
					ecrPrivateRepoDescriber: ecrProvider,
					ecrPublicRepoDescriber:  ecrPublicProvider,
				}, nil
			},
		}

		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(s.log, cfg)
		s.NoError(err)
		s.NotNil(fetcher)

		ecrFetcher, ok := fetcher.(*ECRFetcher)
		s.True(ok)
		s.Equal(ecrProvider, ecrFetcher.PodDescribers[0].Provider)
		s.Equal(ecrPublicProvider, ecrFetcher.PodDescribers[1].Provider)
		s.Equal(kubeclient, ecrFetcher.kubeClient)
		s.Equal(test.expectedRegex[0], ecrFetcher.PodDescribers[0].FilterRegex.String())
		s.Equal(test.expectedRegex[1], ecrFetcher.PodDescribers[1].FilterRegex.String())
	}
}
