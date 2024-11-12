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

package preset

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/elastic-agent-libs/logp"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/condition"
	awsfetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/aws"
	k8sfetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/k8s"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ecr"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/elb"
	"github.com/elastic/cloudbeat/internal/uniqueness"
)

const (
	elbRegexTemplate = "([\\w-]+)-\\d+\\.%s.elb.amazonaws.com"
)

var (
	eksRequiredProcesses = k8sfetchers.ProcessesConfigMap{"kubelet": {ConfigFileArguments: []string{"config"}}}
	eksFsPatterns        = []string{
		"/hostfs/etc/kubernetes/kubelet/kubelet-config.json",
		"/hostfs/var/lib/kubelet/kubeconfig",
	}
)

func NewCisEksFetchers(ctx context.Context, log *logp.Logger, awsConfig aws.Config, ch chan fetching.ResourceInfo, le uniqueness.Manager, k8sClient k8s.Interface, identity *cloud.Identity) registry.FetchersMap {
	log.Infof("Initializing EKS fetchers")
	m := make(registry.FetchersMap)

	if identity != nil {
		log.Info("Initialize aws-related fetchers")
		ecrPrivateProvider := ecr.NewEcrProvider(ctx, log, awsConfig, &awslib.MultiRegionClientFactory[ecr.Client]{})
		privateRepoRegex := fmt.Sprintf(awsfetchers.PrivateRepoRegexTemplate, identity.Account)

		ecrPodDescriber := awsfetchers.PodDescriber{
			FilterRegex: regexp.MustCompile(privateRepoRegex),
			Provider:    ecrPrivateProvider,
		}

		ecrFetcher := awsfetchers.NewEcrFetcher(log, ch, k8sClient, ecrPodDescriber)
		m[fetching.EcrType] = registry.RegisteredFetcher{Fetcher: ecrFetcher, Condition: []fetching.Condition{condition.NewIsLeader(le)}}

		elbProvider := elb.NewElbProvider(ctx, log, identity.Account, awsConfig, &awslib.MultiRegionClientFactory[elb.Client]{})
		loadBalancerRegex := fmt.Sprintf(elbRegexTemplate, awsConfig.Region)
		elbFetcher := awsfetchers.NewElbFetcher(log, ch, k8sClient, elbProvider, identity, loadBalancerRegex)
		m[fetching.ElbType] = registry.RegisteredFetcher{Fetcher: elbFetcher, Condition: []fetching.Condition{condition.NewIsLeader(le)}}
	}

	fsFetcher := k8sfetchers.NewFsFetcher(log, ch, eksFsPatterns)
	m[fetching.FileSystemType] = registry.RegisteredFetcher{Fetcher: fsFetcher}

	procFetcher := k8sfetchers.NewProcessFetcher(log, ch, eksRequiredProcesses)
	m[fetching.ProcessType] = registry.RegisteredFetcher{Fetcher: procFetcher}

	kubeFetcher := k8sfetchers.NewKubeFetcher(log, ch, k8sClient)
	m[fetching.KubeAPIType] = registry.RegisteredFetcher{Fetcher: kubeFetcher, Condition: []fetching.Condition{condition.NewIsLeader(le)}}
	return m
}
