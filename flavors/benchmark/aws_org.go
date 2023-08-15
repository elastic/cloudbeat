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

	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/fetching/factory"
	"github.com/elastic/cloudbeat/resources/fetching/registry"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

type AWSOrg struct {
	IdentityProvider awslib.IdentityProviderGetter
	AccountProvider  awslib.AccountProviderAPI
}

func (a *AWSOrg) Initialize(ctx context.Context, log *logp.Logger, cfg *config.Config, ch chan fetching.ResourceInfo) (registry.Registry, dataprovider.CommonDataProvider, error) {
	if err := a.checkDependencies(); err != nil {
		return nil, nil, err
	}

	// TODO: make this mock-able
	awsConfig, err := aws.InitializeAWSConfig(cfg.CloudConfig.Aws.Cred)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize AWS credentials: %w", err)
	}

	awsIdentity, err := a.IdentityProvider.GetIdentity(ctx, awsConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get AWS identity: %w", err)
	}

	accounts, err := a.getAwsAccounts(ctx, awsConfig, awsIdentity)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get AWS accounts: %w", err)
	}

	return registry.NewRegistry(
			log,
			factory.NewCisAwsOrganizationFactory(ctx, log, ch, accounts),
		), cloud.NewDataProvider(
			cloud.WithLogger(log),
			cloud.WithAccount(*awsIdentity),
		), nil
}

func (a *AWSOrg) getAwsAccounts(ctx context.Context, initialCfg awssdk.Config, rootIdentity *cloud.Identity) ([]factory.AwsAccount, error) {
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

	accountIdentities, err := a.AccountProvider.ListAccounts(ctx, rootCfg)
	if err != nil {
		return nil, err
	}

	var accounts []factory.AwsAccount
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

		accounts = append(accounts, factory.AwsAccount{
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

func (a *AWSOrg) Run(context.Context) error { return nil }
func (a *AWSOrg) Stop()                     {}
