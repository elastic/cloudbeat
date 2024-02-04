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

package benchmark

import (
	"context"
	"errors"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/elastic/beats/v7/x-pack/libbeat/common/aws"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/flavors/benchmark/builder"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/preset"
	"github.com/elastic/cloudbeat/internal/resources/fetching/registry"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type AWSOrg struct {
	IdentityProvider awslib.IdentityProviderGetter
	AccountProvider  awslib.AccountProviderAPI
}

func (a *AWSOrg) NewBenchmark(ctx context.Context, log *logp.Logger, cfg *config.Config) (builder.Benchmark, error) {
	resourceCh := make(chan fetching.ResourceInfo, resourceChBufferSize)
	reg, bdp, _, err := a.initialize(ctx, log, cfg, resourceCh)
	if err != nil {
		return nil, err
	}

	return builder.New(
		builder.WithBenchmarkDataProvider(bdp),
	).Build(ctx, log, cfg, resourceCh, reg)
}

//revive:disable-next-line:function-result-limit
func (a *AWSOrg) initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, dataprovider.IdProvider, error) {
	if err := a.checkDependencies(); err != nil {
		return nil, nil, nil, err
	}

	// TODO: make this mock-able
	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.Aws.Cred)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	awsIdentity, err := a.IdentityProvider.GetIdentity(ctx, awsConfig)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get AWS identity: %w", err)
	}

	cache := make(map[string]registry.FetchersMap)
	reg := registry.NewRegistry(log, registry.WithUpdater(
		func() (registry.FetchersMap, error) {
			accounts, err := a.getAwsAccounts(ctx, log, awsConfig, awsIdentity)
			if err != nil {
				return nil, fmt.Errorf("failed to get AWS accounts: %w", err)
			}

			fm := preset.NewCisAwsOrganizationFetchers(ctx, log, ch, accounts, cache)
			m := make(registry.FetchersMap)
			for accountId, fetchersMap := range fm {
				for key, fetcher := range fetchersMap {
					m[fmt.Sprintf("%s-%s", accountId, key)] = fetcher
				}
			}

			return m, nil
		}))

	return reg, cloud.NewDataProvider(), nil, nil
}

func (a *AWSOrg) getAwsAccounts(ctx context.Context, log *logp.Logger, initialCfg awssdk.Config, rootIdentity *cloud.Identity) ([]preset.AwsAccount, error) {
	const (
		rootRole   = "cloudbeat-root"
		memberRole = "cloudbeat-securityaudit"
	)

	rootCfg := assumeRole(
		sts.NewFromConfig(initialCfg),
		initialCfg,
		fmtIAMRole(rootIdentity.Account, rootRole),
	)
	stsClient := sts.NewFromConfig(rootCfg)

	accountIdentities, err := a.AccountProvider.ListAccounts(ctx, log, rootCfg)
	if err != nil {
		return nil, err
	}

	accounts := make([]preset.AwsAccount, 0, len(accountIdentities))
	for _, identity := range accountIdentities {
		var memberCfg awssdk.Config
		if identity.Account == rootIdentity.Account {
			memberCfg = rootCfg
		} else {
			memberCfg = assumeRole(
				stsClient,
				rootCfg,
				fmtIAMRole(identity.Account, memberRole),
			)
		}

		accounts = append(accounts, preset.AwsAccount{
			Identity: identity,
			Config:   memberCfg,
		})
	}
	return accounts, nil
}

func (a *AWSOrg) checkDependencies() error {
	if a.IdentityProvider == nil {
		return errors.New("aws identity provider is uninitialized")
	}
	if a.AccountProvider == nil {
		return errors.New("aws account provider is uninitialized")
	}
	return nil
}

func assumeRole(client *sts.Client, cfg awssdk.Config, arn string) awssdk.Config {
	cfg.Credentials = awssdk.NewCredentialsCache(stscreds.NewAssumeRoleProvider(client, arn))
	return cfg
}

func fmtIAMRole(account string, role string) string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", account, role)
}
