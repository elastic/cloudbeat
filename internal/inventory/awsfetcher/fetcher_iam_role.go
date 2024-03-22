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

package awsfetcher

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type IamRoleFetcher struct {
	logger      *logp.Logger
	provider    IamRoleProvider
	AccountId   string
	AccountName string
}

type IamRoleProvider interface {
	ListRoles(ctx context.Context) ([]*iam.Role, error)
}

var iamRoleClassification = inventory.AssetClassification{
	Category:    inventory.CategoryIdentity,
	SubCategory: inventory.SubCategoryCloudProviderAccount,
	Type:        inventory.TypeServiceAccount,
	SubType:     inventory.SubTypeIAM,
}

func newIamRoleFetcher(logger *logp.Logger, identity *cloud.Identity, cfg aws.Config) inventory.AssetFetcher {
	provider := iam.NewIAMProvider(logger, cfg, &awslib.MultiRegionClientFactory[iam.AccessAnalyzerClient]{})
	return &IamRoleFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (i *IamRoleFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	i.logger.Info("Fetching IAM Roles")
	defer i.logger.Info("Fetching IAM Roles - Finished")

	users, err := i.provider.ListRoles(ctx)
	if err != nil {
		i.logger.Errorf("Could not list roles: %v", err)
		if len(users) == 0 {
			return
		}
	}

	for _, role := range users {
		if role == nil {
			continue
		}

		assetChannel <- inventory.NewAssetEvent(
			iamRoleClassification,
			pointers.Deref(role.Arn),
			pointers.Deref(role.RoleName),

			inventory.WithRawAsset(*role),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   awslib.GlobalRegion,
				Account: inventory.AssetCloudAccount{
					Id:   i.AccountId,
					Name: i.AccountName,
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS IAM",
				},
			}),
		)
	}
}
