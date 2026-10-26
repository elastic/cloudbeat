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

package assetinventory

import (
	"context"
	"errors"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/awsfetcher"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

func (s *strategy) getInitialAWSConfig(ctx context.Context, cfg *config.Config) (*awssdk.Config, error) {
	if cfg.CloudConfig.Aws.CloudConnectors {
		return awslib.InitializeAWSConfigCloudConnectors(ctx, cfg.CloudConfig.Aws)
	}

	return awslib.InitializeAWSConfig(cfg.CloudConfig.Aws.Cred, s.logger.Logger)
}

// getAWSIdentity makes the aws.Config from the configuration and gets the identity of the credentials.
func (s *strategy) getAWSIdentity(ctx context.Context) (*awssdk.Config, *cloud.Identity, error) {
	awsConfig, err := s.getInitialAWSConfig(ctx, s.cfg)
	if err != nil {
		return nil, nil, err
	}

	idProvider := awslib.IdentityProvider{Logger: s.logger}
	awsIdentity, err := idProvider.GetIdentity(ctx, *awsConfig)
	if err != nil {
		return nil, nil, err
	}

	return awsConfig, awsIdentity, nil
}

func (s *strategy) initAwsFetchers(ctx context.Context, statusHandler statushandler.StatusHandlerAPI) ([]inventory.AssetFetcher, error) {
	s.logger.Infof("initializing asset inventory aws (cloud connectors %t)", s.cfg.CloudConfig.Aws.CloudConnectors)

	awsConfig, awsIdentity, err := s.getAWSIdentity(ctx)
	if err != nil && s.cfg.CloudConfig.Aws.Cred.DefaultRegion == "" {
		// If the default region is the commercial region (us-east-1), GovCloud credentials fail.
		// The CSPM benchmark uses the same fallback (see benchmark.AWS and benchmark.AWSOrg).
		s.logger.Warnf("failed to initialize identity (%v); retrying to check AWS Gov Cloud regions", err)
		s.cfg.CloudConfig.Aws.Cred.DefaultRegion = awslib.DefaultGovRegion
		firstErr := err
		awsConfig, awsIdentity, err = s.getAWSIdentity(ctx)
		if err != nil {
			// Keep the first error. It can show the real cause, for example credentials that are not valid.
			err = errors.Join(firstErr, fmt.Errorf("retry in region %s: %w", awslib.DefaultGovRegion, err))
		}
	}
	if err != nil {
		return nil, err
	}

	orgIAMRoleNamesProvider := getOrgIAMRoleNamesProvider(s.cfg.CloudConfig.Aws)

	// Early exit if we're scanning the entire account.
	if s.cfg.CloudConfig.Aws.AccountType == config.SingleAccount {
		return awsfetcher.New(ctx, s.logger, awsIdentity, *awsConfig, statusHandler), nil
	}

	// Assume audit roles per selected account and generate fetchers for them
	// The roles are in the partition of the caller identity (for example "aws-us-gov").
	rootRoleConfig := assumeRole(
		sts.NewFromConfig(*awsConfig),
		*awsConfig,
		awslib.IAMRoleARN(awsIdentity.Partition, awsIdentity.Account, orgIAMRoleNamesProvider.RootRoleName()),
	)
	accountProvider := &awslib.AccountProvider{}
	accountIdentities, err := accountProvider.ListAccounts(ctx, s.logger, rootRoleConfig)
	if err != nil {
		return nil, err
	}
	var fetchers []inventory.AssetFetcher
	stsClient := sts.NewFromConfig(rootRoleConfig)
	for _, identity := range accountIdentities {
		assumedRoleConfig := assumeRole(
			stsClient,
			rootRoleConfig,
			awslib.IAMRoleARN(identity.Partition, identity.Account, orgIAMRoleNamesProvider.MemberRoleName()),
		)
		if ok := awslib.CredentialsValid(ctx, assumedRoleConfig, s.logger); !ok {
			// role does not exist, skip identity/account
			s.logger.Infof("Skipping identity on purpose %+v", identity)
			continue
		}
		accountFetchers := awsfetcher.New(ctx, s.logger, &identity, assumedRoleConfig, statusHandler)
		fetchers = append(fetchers, accountFetchers...)
	}

	return fetchers, nil
}

func assumeRole(client stscreds.AssumeRoleAPIClient, cfg awssdk.Config, arn string) awssdk.Config {
	cfg.Credentials = awssdk.NewCredentialsCache(stscreds.NewAssumeRoleProvider(client, arn))
	return cfg
}

func getOrgIAMRoleNamesProvider(cfg config.AwsConfig) awslib.OrgIAMRoleNamesProvider {
	if cfg.CloudConnectors {
		return awslib.BenchmarkOrgIAMRoleNamesProvider{} // for reusability with CSPM when cloud connectors are enabled.
	}
	return awslib.AssetDiscoveryOrgIAMRoleNamesProvider{}
}
