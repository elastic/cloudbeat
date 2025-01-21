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
	"fmt"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/awsfetcher"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

const (
	rootRole   = "cloudbeat-asset-inventory-root"
	memberRole = "cloudbeat-asset-inventory-securityaudit"
)

func (s *strategy) initAwsFetchers(ctx context.Context) ([]inventory.AssetFetcher, error) {
	awsConfig, err := awslib.InitializeAWSConfig(s.cfg.CloudConfig.Aws.Cred)
	if err != nil {
		return nil, err
	}

	idProvider := awslib.IdentityProvider{Logger: s.logger}
	awsIdentity, err := idProvider.GetIdentity(ctx, *awsConfig)
	if err != nil {
		return nil, err
	}

	// Early exit if we're scanning the entire account.
	if s.cfg.CloudConfig.Aws.AccountType == config.SingleAccount {
		return awsfetcher.New(ctx, s.logger, awsIdentity, *awsConfig), nil
	}

	// Assume audit roles per selected account and generate fetchers for them
	rootRoleConfig := assumeRole(
		sts.NewFromConfig(*awsConfig),
		*awsConfig,
		fmtIAMRole(awsIdentity.Account, rootRole),
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
			fmtIAMRole(identity.Account, memberRole),
		)
		if ok := tryListingBuckets(ctx, s.logger, assumedRoleConfig); !ok {
			// role does not exist, skip identity/account
			s.logger.Infof("Skipping identity on purpose %+v", identity)
			continue
		}
		accountFetchers := awsfetcher.New(ctx, s.logger, &identity, assumedRoleConfig)
		fetchers = append(fetchers, accountFetchers...)
	}

	return fetchers, nil
}

func tryListingBuckets(ctx context.Context, log *clog.Logger, roleConfig awssdk.Config) bool {
	s3Client := s3.NewFromConfig(roleConfig)
	_, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{MaxBuckets: pointers.Ref(int32(1))})
	if err == nil {
		return true
	}
	if !strings.Contains(err.Error(), "not authorized to perform: sts:AssumeRole") {
		log.Errorf("Expected a 403 autorization error, but got: %v", err)
	}
	return false
}

func assumeRole(client stscreds.AssumeRoleAPIClient, cfg awssdk.Config, arn string) awssdk.Config {
	cfg.Credentials = awssdk.NewCredentialsCache(stscreds.NewAssumeRoleProvider(client, arn))
	return cfg
}

func fmtIAMRole(account string, role string) string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", account, role)
}
