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
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"testing"

	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
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
access_key_id: key
secret_access_key: secret
session_token: session
default_region: us1-east
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

		mockedConfigGetter := &config.MockAwsConfigProvider{}
		mockedConfigGetter.EXPECT().
			InitializeAWSConfig(mock.Anything, mock.Anything).
			Call.
			Return(func(ctx context.Context, config aws.ConfigAWS) awssdk.Config {
				return CreateSdkConfig(config, "us1-east")
			},
				func(ctx context.Context, config aws.ConfigAWS) error {
					return nil
				},
			)

		identity := awslib.Identity{
			Account: &test.account,
		}
		identityProvider := &awslib.MockedIdentityProviderGetter{}
		identityProvider.EXPECT().GetIdentity(mock.Anything).Return(&identity, nil)

		factory := &ECRFactory{
			KubernetesProvider: mockedKubernetesClientGetter,
			AwsConfigProvider:  mockedConfigGetter,
			IdentityProvider: func(cfg awssdk.Config) awslib.IdentityProviderGetter {
				return identityProvider
			},
		}

		cfg, err := agentconfig.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(s.log, cfg, nil)
		s.NoError(err)
		s.NotNil(fetcher)

		ecrFetcher, ok := fetcher.(*ECRFetcher)
		s.True(ok)
		s.Equal(kubeclient, ecrFetcher.kubeClient)
		s.Equal(test.expectedRegex[0], ecrFetcher.PodDescribers[0].FilterRegex.String())
		s.Equal(test.expectedRegex[1], ecrFetcher.PodDescribers[1].FilterRegex.String())
	}
}

func CreateSdkConfig(config aws.ConfigAWS, region string) awssdk.Config {
	awsConfig := awssdk.NewConfig()
	awsCredentials := awssdk.Credentials{
		AccessKeyID:     config.AccessKeyID,
		SecretAccessKey: config.SecretAccessKey,
		SessionToken:    config.SessionToken,
	}

	awsConfig.Credentials = awssdk.StaticCredentialsProvider{
		Value: awsCredentials,
	}
	awsConfig.Region = region
	return *awsConfig
}
