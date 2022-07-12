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
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/providers"
	"regexp"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/docker/distribution/context"
	"github.com/elastic/cloudbeat/resources/fetching"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	ECRType = "aws-ecr"
)

func init() {
	fetchersManager.Factories.RegisterFactory(ECRType, &ECRFactory{
		KubernetesProvider: providers.KubernetesProvider{},
		IdentityProvider:   awslib.GetIdentityClient,
	})
}

type ECRFactory struct {
	KubernetesProvider providers.KubernetesClientGetter
	IdentityProvider   func(cfg awssdk.Config) awslib.IdentityProviderGetter
}

func (f *ECRFactory) Create(log *logp.Logger, c *config.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting ECRFactory.Create")

	cfg := ECRFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}
	return f.CreateFrom(log, cfg, ch)
}

func (f *ECRFactory) CreateFrom(log *logp.Logger, cfg ECRFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	awsConfig, err := aws.InitializeAWSConfig(cfg.AwsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	ecrPrivateProvider := awslib.NewEcrProvider(awsConfig)
	ecrPublicProvider := awslib.NewEcrPublicProvider()
	kubeClient, err := f.KubernetesProvider.GetClient(cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes client: %w", err)
	}

	ctx := context.Background()
	identityProvider := f.IdentityProvider(awsConfig)
	identity, err := identityProvider.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve user identity for ECR fetcher: %w", err)
	}

	privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, *identity.Account, awsConfig.Region)

	privateECRExecutor := PodDescriber{
		FilterRegex: regexp.MustCompile(privateRepoRegex),
		Provider:    ecrPrivateProvider,
	}
	publicECRExecutor := PodDescriber{
		FilterRegex: regexp.MustCompile(PublicRepoRegex),
		Provider:    ecrPublicProvider,
	}

	fe := &ECRFetcher{
		log:        log,
		cfg:        cfg,
		kubeClient: kubeClient,
		PodDescribers: []PodDescriber{
			privateECRExecutor,
			publicECRExecutor,
		},
		resourceCh: ch,
	}
	return fe, nil
}
