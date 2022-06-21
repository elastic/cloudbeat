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
	"regexp"

	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
	common "github.com/elastic/elastic-agent-libs/config"
)

const (
	ELBType = "aws-elb"
)

func init() {

	manager.Factories.ListFetcherFactory(ELBType,
		&ELBFactory{
			extraElements: getElbExtraElements,
		},
	)
}

type ELBFactory struct {
	extraElements func() (elbExtraElements, error)
}

type elbExtraElements struct {
	balancerDescriber      awslib.ELBLoadBalancerDescriber
	awsConfig              awslib.Config
	kubernetesClientGetter providers.KubernetesClientGetter
}

func (f *ELBFactory) Create(log *logp.Logger, c *common.C) (fetching.Fetcher, error) {
	log.Debug("Starting ELBFactory.Create")

	cfg := ELBFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}
	elements, err := f.extraElements()
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(log, cfg, elements)
}

func getElbExtraElements() (elbExtraElements, error) {
	awsConfigProvider := awslib.ConfigProvider{}
	awsConfig, err := awsConfigProvider.GetConfig()
	if err != nil {
		return elbExtraElements{}, err
	}
	elb := awslib.NewELBProvider(awsConfig.Config)
	kubeGetter := providers.KubernetesProvider{}

	return elbExtraElements{
		balancerDescriber:      elb,
		awsConfig:              awsConfig,
		kubernetesClientGetter: kubeGetter,
	}, err
}

func (f *ELBFactory) CreateFrom(log *logp.Logger, cfg ELBFetcherConfig, elements elbExtraElements) (fetching.Fetcher, error) {
	loadBalancerRegex := fmt.Sprintf(ELBRegexTemplate, elements.awsConfig.Config.Region)
	kubeClient, err := elements.kubernetesClientGetter.GetClient(cfg.Kubeconfig, kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, fmt.Errorf("could not initate Kubernetes: %w", err)
	}

	return &ELBFetcher{
		log:             log,
		elbProvider:     elements.balancerDescriber,
		cfg:             cfg,
		kubeClient:      kubeClient,
		lbRegexMatchers: []*regexp.Regexp{regexp.MustCompile(loadBalancerRegex)},
	}, nil
}
