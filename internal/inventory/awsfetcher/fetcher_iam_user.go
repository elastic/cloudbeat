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

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
)

type iamUserFetcher struct {
	logger      *logp.Logger
	provider    iamUserProvider
	AccountId   string
	AccountName string
}

type iamUserProvider interface {
	GetUsers(ctx context.Context) ([]awslib.AwsResource, error)
}

var iamUserClassification = inventory.AssetClassification{
	Category:    inventory.CategoryIdentity,
	SubCategory: inventory.SubCategoryCloudProviderAccount,
	Type:        inventory.TypeUser,
	SubType:     inventory.SubTypeIAM,
}

func newIamUserFetcher(logger *logp.Logger, identity *cloud.Identity, provider iamUserProvider) inventory.AssetFetcher {
	return &iamUserFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (i *iamUserFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	i.logger.Info("Fetching IAM Users")
	defer i.logger.Info("Fetching IAM Users - Finished")

	users, err := i.provider.GetUsers(ctx)
	if err != nil {
		i.logger.Errorf("Could not list users: %v", err)
		if len(users) == 0 {
			return
		}
	}

	for _, resource := range users {
		if resource == nil {
			continue
		}

		user, ok := resource.(iam.User)
		if !ok {
			i.logger.Errorf("Could not get info about user: %s", resource.GetResourceArn())
			continue
		}

		assetChannel <- inventory.NewAssetEvent(
			iamUserClassification,
			inventory.Identifiers(inventory.Arns(user.GetResourceArn()), inventory.Ids(user.UserId)),
			user.GetResourceName(),

			inventory.WithRawAsset(user),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   user.GetRegion(),
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
