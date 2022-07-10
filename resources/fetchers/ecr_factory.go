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
	"regexp"

	"github.com/docker/distribution/context"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	ECRType = "aws-ecr"
)

func init() {
	fetchersManager.Factories.SetFetcherFactory(ECRType, &ECRFactory{extraElements: getEcrExtraElements})
}

type ECRFactory struct {
	extraElements func() (ecrExtraElements, error)
}

type ecrExtraElements struct {
	awsConfig               awslib.Config
	kubernetesClientGetter  providers.KubernetesClientGetter
	identityProviderGetter  awslib.IdentityProviderGetter
	ecrPrivateRepoDescriber awslib.EcrRepositoryDescriber
	ecrPublicRepoDescriber  awslib.EcrRepositoryDescriber
}

func (f *ECRFactory) Create(log *logp.Logger, c *config.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting ECRFactory.Create")

	cfg := ECRFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}
	elements, err := f.extraElements()
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(log, cfg, elements, ch)
}

func getEcrExtraElements() (ecrExtraElements, error) {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig, err := awsConfigProvider.GetConfig()
	if err != nil {
		return ecrExtraElements{}, err
	}
	ecrPrivateProvider := awslib.NewEcrProvider(awsConfig.Config)
	ecrPublicProvider := awslib.NewEcrPublicProvider()
	identityProvider := awslib.NewAWSIdentityProvider(awsConfig.Config)
	kubeGetter := providers.KubernetesProvider{}

	extraElements := ecrExtraElements{
		awsConfig:               awsConfig,
		kubernetesClientGetter:  kubeGetter,
		identityProviderGetter:  identityProvider,
		ecrPrivateRepoDescriber: ecrPrivateProvider,
		ecrPublicRepoDescriber:  ecrPublicProvider,
	}

	return extraElements, nil
}

func (f *ECRFactory) CreateFrom(log *logp.Logger, cfg ECRFetcherConfig, elements ecrExtraElements, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	ctx := context.Background()
	identity, err := elements.identityProviderGetter.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve user identity for ECR fetcher: %w", err)
	}

	privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, *identity.Account, elements.awsConfig.Config.Region)
	kubeClient, err := elements.kubernetesClientGetter.GetClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes client: %w", err)
	}

	privateECRExecutor := PodDescriber{
		FilterRegex: regexp.MustCompile(privateRepoRegex),
		Provider:    elements.ecrPrivateRepoDescriber,
	}
	publicECRExecutor := PodDescriber{
		FilterRegex: regexp.MustCompile(PublicRepoRegex),
		Provider:    elements.ecrPublicRepoDescriber,
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
