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
	"github.com/elastic/cloudbeat/resources/conditions"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/ecr"
	"github.com/elastic/cloudbeat/resources/providers/awslib/elb"
	"github.com/elastic/cloudbeat/uniqueness"
	"github.com/elastic/elastic-agent-libs/logp"
	k8s "k8s.io/client-go/kubernetes"
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

func NewCisEksFactory(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo, le uniqueness.Manager, k8sClient k8s.Interface) (FetchersMap, error) {
	log.Infof("Initializing EKS fetchers")

	m := make(FetchersMap)
	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.AwsCred)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	identity, err := awslib.GetIdentityClient(awsConfig).GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get cloud indentity: %w", err)
	}

	fsFetcher := fetchers.NewFsFetcher(log, ch, eksFsPatterns)
	m[fetching.FileSystemType] = RegisteredFetcher{Fetcher: fsFetcher}

	procFetcher := fetchers.NewProcessFetcher(log, ch, eksRequiredProcesses)
	m[fetching.ProcessType] = RegisteredFetcher{Fetcher: procFetcher}

	kubeFetcher := fetchers.NewKubeFetcher(log, ch, k8sClient)
	m[fetching.KubeAPIType] = RegisteredFetcher{Fetcher: kubeFetcher, Condition: []fetching.Condition{conditions.NewLeaseFetcherCondition(log, le)}}

	ecrPrivateProvider := ecr.NewEcrProvider(log, awsConfig, &awslib.MultiRegionClientFactory[ecr.Client]{})
	privateRepoRegex := fmt.Sprintf(fetchers.PrivateRepoRegexTemplate, *identity.Account)

	ecrPodDescriber := fetchers.PodDescriber{
		FilterRegex: regexp.MustCompile(privateRepoRegex),
		Provider:    ecrPrivateProvider,
	}

	ecrFetcher := fetchers.NewEcrFetcher(log, ch, k8sClient, ecrPodDescriber)
	m[fetching.EcrType] = RegisteredFetcher{Fetcher: ecrFetcher, Condition: []fetching.Condition{conditions.NewLeaseFetcherCondition(log, le)}}

	elbProvider := elb.NewElbProvider(awsConfig)
	loadBalancerRegex := fmt.Sprintf(elbRegexTemplate, awsConfig.Region)
	elbFetcher := fetchers.NewElbFetcher(log, ch, k8sClient, elbProvider, identity, loadBalancerRegex)
	m[fetching.ElbType] = RegisteredFetcher{Fetcher: elbFetcher, Condition: []fetching.Condition{conditions.NewLeaseFetcherCondition(log, le)}}

	return m, nil
}
