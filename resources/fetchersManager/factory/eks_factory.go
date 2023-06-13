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

package factory

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/ecr"
	"github.com/elastic/cloudbeat/resources/providers/awslib/elb"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	"regexp"
)

const (
	elbRegexTemplate = "([\\w-]+)-\\d+\\.%s.elb.amazonaws.com"
)

var (
	eksRequiredProcesses = fetchers.ProcessesConfigMap{"kubelet": {ConfigFileArguments: []string{"config"}}}
	eksFsPatterns        = []string{
		"/hostfs/etc/kubernetes/kubelet/kubelet-config.json",
		"/hostfs/var/lib/kubelet/kubeconfig"}
)

func NewCisEksFactory(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (FetchersMap, error) {
	log.Infof("Initializing EKS fetchers")

	m := make(FetchersMap)
	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.AwsCred)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	kubeClient, err := providers.GetK8sClient(log, cfg.KubeConfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes client: %w", err)
	}

	identity, err := awslib.GetIdentityClient(awsConfig).GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get cloud indentity: %w", err)
	}

	fsFetcher := fetchers.NewFsFetcher(log, ch, eksFsPatterns)
	m[fetching.FileSystemType] = fsFetcher

	procFetcher := fetchers.NewProcessFetcher(log, ch, eksRequiredProcesses)
	m[fetching.ProcessType] = procFetcher

	kubeFetcher := fetchers.NewKubeFetcher(log, ch, kubeClient)
	m[fetching.KubeAPIType] = kubeFetcher

	ecrPrivateProvider := ecr.NewEcrProvider(log, awsConfig, &awslib.MultiRegionClientFactory[ecr.Client]{})
	privateRepoRegex := fmt.Sprintf(fetchers.PrivateRepoRegexTemplate, *identity.Account)

	ecrPodDescriber := fetchers.PodDescriber{
		FilterRegex: regexp.MustCompile(privateRepoRegex),
		Provider:    ecrPrivateProvider,
	}

	ecrFetcher := fetchers.NewEcrFetcher(log, ch, kubeClient, ecrPodDescriber)
	m[fetching.EcrType] = ecrFetcher

	elbProvider := elb.NewElbProvider(awsConfig)
	loadBalancerRegex := fmt.Sprintf(elbRegexTemplate, awsConfig.Region)
	elbFetcher := fetchers.NewElbFetcher(log, ch, kubeClient, elbProvider, identity, loadBalancerRegex)
	m[fetching.ElbType] = elbFetcher

	return m, nil
}
