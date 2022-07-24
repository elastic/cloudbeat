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
	"github.com/stretchr/testify/mock"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"testing"

	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/suite"
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
		account       string
		expectedRegex string
	}{
		{
			`
name: aws-elb
access_key_id: key
secret_access_key: secret
session_token: session
default_region: us2-east
`,
			"us2-east",
			"my-account",
			"([\\w-]+)-\\d+\\.us2-east.elb.amazonaws.com",
		},
	}

	for _, test := range tests {

		kubeclient := k8sfake.NewSimpleClientset()
		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		identity := awslib.Identity{
			Account: &test.account,
		}
		identityProvider := &awslib.MockIdentityProviderGetter{}
		identityProvider.EXPECT().GetIdentity(mock.Anything).Return(&identity, nil)

		mockedConfigGetter := &config.MockAwsConfigProvider{}
		mockedConfigGetter.EXPECT().
			InitializeAWSConfig(mock.Anything, mock.Anything).
			Call.
			Return(func(ctx context.Context, config aws.ConfigAWS) awssdk.Config {
				return CreateSdkConfig(config, "us2-east")
			},
				func(ctx context.Context, config aws.ConfigAWS) error {
					return nil
				},
			)
		factory := &ELBFactory{
			KubernetesProvider: mockedKubernetesClientGetter,
			IdentityProvider: func(cfg awssdk.Config) awslib.IdentityProviderGetter {
				return identityProvider
			},
			AwsConfigProvider: mockedConfigGetter,
		}

		cfg, err := agentconfig.NewConfigFrom(test.config)
		s.NoError(err)

		fetcher, err := factory.Create(s.log, cfg, nil)
		s.NoError(err)
		s.NotNil(fetcher)

		elbFetcher, ok := fetcher.(*ELBFetcher)
		s.True(ok)
		s.Equal(test.expectedRegex, elbFetcher.lbRegexMatchers[0].String())
		s.Equal(kubeclient, elbFetcher.kubeClient)
		s.Equal("key", elbFetcher.cfg.AwsConfig.AccessKeyID)
		s.Equal("secret", elbFetcher.cfg.AwsConfig.SecretAccessKey)
		s.Equal("session", elbFetcher.cfg.AwsConfig.SessionToken)
	}
}
