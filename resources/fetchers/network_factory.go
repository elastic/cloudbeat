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
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/fetchersManager"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/ec2"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetching"
)

func init() {
	fetchersManager.Factories.RegisterFactory(fetching.EC2NetworkingType, &EC2NetworkFactory{
		CrossRegionUtil:  &awslib.MultiRegionClientFactory[ec2.ElasticCompute]{},
		IdentityProvider: awslib.GetIdentityClient,
	})
}

type EC2NetworkFactory struct {
	CrossRegionUtil  awslib.CrossRegionFactory[ec2.ElasticCompute]
	IdentityProvider func(cfg awssdk.Config) awslib.IdentityProviderGetter
}

func (f *EC2NetworkFactory) Create(log *logp.Logger, c *agentconfig.C, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	log.Debug("Starting EC2NetworkFactory.Create")

	cfg := ACLFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(log, cfg, ch)
}

func (f *EC2NetworkFactory) CreateFrom(log *logp.Logger, cfg ACLFetcherConfig, ch chan fetching.ResourceInfo) (fetching.Fetcher, error) {
	ctx := context.Background()
	awsConfig, err := aws.InitializeAWSConfig(cfg.AwsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	identityProvider := f.IdentityProvider(awsConfig)
	identity, err := identityProvider.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get cloud indentity: %w", err)
	}

	provider := ec2.NewCrossEC2Provider(log, *identity.Account, awsConfig, f.CrossRegionUtil)

	return &NetworkFetcher{
		log:           log,
		cfg:           cfg,
		provider:      provider,
		cloudIdentity: identity,
		resourceCh:    ch,
	}, nil
}
