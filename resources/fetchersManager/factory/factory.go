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
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/uniqueness"
	"github.com/elastic/elastic-agent-libs/logp"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/elastic/cloudbeat/resources/fetching"
)

type RegisteredFetcher struct {
	Fetcher   fetching.Fetcher
	Condition []fetching.Condition
}

type FetchersMap map[string]RegisteredFetcher

// NewFactory Creates a new factory based on the benchmark name
func NewFactory(
	ctx context.Context,
	log *logp.Logger,
	cfg *config.Config,
	ch chan fetching.ResourceInfo,
	le uniqueness.Manager,
	k8sClient k8s.Interface,
	identityProvider func(cfg awssdk.Config) awslib.IdentityProviderGetter,
) (FetchersMap, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	var awsIdentity *awslib.Identity
	var awsConfig awssdk.Config
	var err error
	if cfg.CloudConfig != (config.CloudConfig{}) && cfg.CloudConfig.AwsCred != (aws.ConfigAWS{}) {
		log.Infof("NewFactory Using AWS credentials from config")
		awsConfig, awsIdentity, err = AdaptAwsConfig(ctx, cfg, identityProvider)
		if err != nil {
			return nil, err
		}
	}

	switch cfg.Benchmark {
	case config.CIS_AWS:
		return NewCisAwsFactory(log, awsConfig, ch, awsIdentity)
	case config.CIS_K8S:
		return NewCisK8sFactory(log, cfg, ch, le, k8sClient)
	case config.CIS_EKS:
		return NewCisEksFactory(log, awsConfig, ch, le, k8sClient, awsIdentity)
	}

	return nil, fmt.Errorf("benchmark %s is not supported, no fetchers to return", cfg.Benchmark)
}

func AdaptAwsConfig(ctx context.Context, cfg *config.Config, identityProvider func(cfg awssdk.Config) awslib.IdentityProviderGetter) (awssdk.Config, *awslib.Identity, error) {
	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.AwsCred)
	if err != nil {
		return awssdk.Config{}, nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	IdentityProviderGetter, err := identityProvider(awsConfig).GetIdentity(ctx)
	if err != nil {
		return awssdk.Config{}, nil, fmt.Errorf("failed to get AWS identity: %w", err)
	}
	return awsConfig, IdentityProviderGetter, nil
}
