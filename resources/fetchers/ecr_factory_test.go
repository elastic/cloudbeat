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
	awsorg "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/aws"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"testing"
)

type EcrFactoryTestSuite struct {
	suite.Suite
	factory fetching.Factory
}

func TestEcrFactoryTestSuite(t *testing.T) {
	suite.Run(t, new(EcrFactoryTestSuite))
}

func (s *EcrFactoryTestSuite) SetupTest() {

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
				"^my-account\\.dkr\\.ecr\\.us1-east\\.amazonaws\\.com\\/([-\\w]+)[:,@]?",
				"public\\.ecr\\.aws\\/\\w+\\/([\\w-]+)\\:?",
			},
		},
	}

	for _, test := range tests {

		kubeclient := k8sfake.NewSimpleClientset()
		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		identity := aws.Identity{
			Account: &test.account,
		}
		identityProvider := &aws.MockedIdentityProviderGetter{}
		identityProvider.EXPECT().GetIdentity(mock.Anything).Return(&identity, nil)
		awsCred := aws.Config{Config: awsorg.Config{
			Region: test.region,
		}}
		mockedAwsCred := &aws.MockedAwsCredentialsGetter{}
		mockedAwsCred.EXPECT().GetAwsCredentials().Return(awsCred)

		ecrProvider := &aws.MockedEcrRepositoryDescriber{}

		factory := &ECRFactory{
			kubernetesClientGetter: mockedKubernetesClientGetter,
			ecrRepoDescriber:       ecrProvider,
			identityProviderGetter: identityProvider,
		}

		cfg, err := common.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(cfg)
		s.NoError(err)
		s.NotNil(fetcher)

		ecrFetcher, ok := fetcher.(*ECRFetcher)
		s.True(ok)
		s.Equal(ecrProvider, ecrFetcher.ecrProvider)
		s.Equal(kubeclient, ecrFetcher.kubeClient)
		s.Equal(test.expectedRegex[0], ecrFetcher.repoRegexMatchers[0].String())
		s.Equal(test.expectedRegex[1], ecrFetcher.repoRegexMatchers[1].String())
	}
}
