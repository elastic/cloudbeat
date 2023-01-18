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
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"regexp"

	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/fetching"
)

func init() {
	fetchersManager.Factories.RegisterFactory(fetching.ElbType, &ElbFactory{
		KubernetesProvider: providers.KubernetesProvider{},
		IdentityProvider:   awslib.GetIdentityClient,
		AwsConfigProvider:  awslib.ConfigProvider{MetadataProvider: awslib.Ec2MetadataProvider{}},
	})
}

type ElbFactory struct {
	KubernetesProvider providers.KubernetesClientGetter
	IdentityProvider   func(cfg awssdk.Config) awslib.IdentityProviderGetter
	AwsConfigProvider  awslib.ConfigProviderAPI
}

func (f *ElbFactory) Create(log *logp.Logger, c *agentconfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting ElbFactory.Create")

	cfg := ElbFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}
	return f.CreateFrom(log, cfg, ch)
}

func (f *ElbFactory) CreateFrom(log *logp.Logger, cfg ElbFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	awsConfig, err := f.AwsConfigProvider.InitializeAWSConfig(ctx, cfg.AwsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}
	loadBalancerRegex := fmt.Sprintf(elbRegexTemplate, awsConfig.Region)
	kubeClient, err := f.KubernetesProvider.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes: %w", err)
	}

	balancerDescriber := awslib.NewElbProvider(*awsConfig)
	identityProvider := f.IdentityProvider(*awsConfig)
	identity, err := identityProvider.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get cloud indentity: %w", err)
	}

	return &ElbFetcher{
		log:             log,
		elbProvider:     balancerDescriber,
		cloudIdentity:   identity,
		cfg:             cfg,
		kubeClient:      kubeClient,
		lbRegexMatchers: []*regexp.Regexp{regexp.MustCompile(loadBalancerRegex)},
		resourceCh:      ch,
	}, nil
}
