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
	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/iam"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

func init() {
	fetchersManager.Factories.RegisterFactory(fetching.IAMType, &IAMFactory{
		AwsConfigProvider: awslib.ConfigProvider{MetadataProvider: awslib.Ec2MetadataProvider{}},
		IdentityProvider:  awslib.GetIdentityClient,
	})
}

type IAMFactory struct {
	AwsConfigProvider config.AwsConfigProvider
	IdentityProvider  func(cfg awssdk.Config) awslib.IdentityProviderGetter
}

func (f *IAMFactory) Create(log *logp.Logger, c *agentconfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting IAMFactory.Create")

	cfg := IAMFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(log, cfg, ch)
}

func (f *IAMFactory) CreateFrom(log *logp.Logger, cfg IAMFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	ctx := context.Background()
	awsConfig, err := f.AwsConfigProvider.InitializeAWSConfig(ctx, cfg.AwsConfig, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	identityProvider := f.IdentityProvider(awsConfig)
	identity, err := identityProvider.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get cloud indentity: %w", err)
	}
	
	provider := iam.NewIAMProvider(log, awsConfig)

	return &IAMFetcher{
		log:           log,
		cfg:           cfg,
		iamProvider:   provider,
		cloudIdentity: identity,
		resourceCh:    ch,
	}, nil

}
