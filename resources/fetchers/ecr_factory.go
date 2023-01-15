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
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/providers"
	"regexp"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"

	"github.com/docker/distribution/context"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

func init() {
	fetchersManager.Factories.RegisterFactory(fetching.EcrType, &EcrFactory{
		KubernetesProvider: providers.KubernetesProvider{},
		IdentityProvider:   awslib.GetIdentityClient,
		AwsConfigProvider:  awslib.ConfigProvider{MetadataProvider: awslib.Ec2MetadataProvider{}},
	})
}

type EcrFactory struct {
	KubernetesProvider providers.KubernetesClientGetter
	IdentityProvider   func(cfg awssdk.Config) awslib.IdentityProviderGetter
	AwsConfigProvider  awslib.ConfigProviderAPI
}

func (f *EcrFactory) Create(log *logp.Logger, c *agentconfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting EcrFactory.Create")

	cfg := EcrFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}
	return f.CreateFrom(log, cfg, ch)
}

func (f *EcrFactory) CreateFrom(log *logp.Logger, cfg EcrFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	ctx := context.Background()
	awsConfig, err := f.AwsConfigProvider.InitializeAWSConfig(ctx, cfg.AwsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	kubeClient, err := f.KubernetesProvider.GetClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes client: %w", err)
	}

	identityProvider := f.IdentityProvider(awsConfig)
	identity, err := identityProvider.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve user identity for ECR fetcher: %w", err)
	}

	ecrPrivateProvider := awslib.NewEcrProvider()
	privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, *identity.Account)

	ecrPodDescriber := PodDescriber{
		FilterRegex: regexp.MustCompile(privateRepoRegex),
		Provider:    ecrPrivateProvider,
	}

	fe := &EcrFetcher{
		log:          log,
		cfg:          cfg,
		kubeClient:   kubeClient,
		PodDescriber: ecrPodDescriber,
		resourceCh:   ch,
		awsConfig:    awsConfig,
	}
	return fe, nil
}
