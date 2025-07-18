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

	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/rds"
)

type rdsFetcher struct {
	logger      *clog.Logger
	provider    rdsProvider
	AccountId   string
	AccountName string
}

type rdsProvider interface {
	DescribeDBInstances(ctx context.Context) ([]awslib.AwsResource, error)
}

func newRDSFetcher(logger *clog.Logger, identity *cloud.Identity, provider rdsProvider) inventory.AssetFetcher {
	return &rdsFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (s *rdsFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	s.logger.Info("Fetching RDS Instances")
	defer s.logger.Info("Fetching RDS Instances - Finished")

	awsResources, err := s.provider.DescribeDBInstances(ctx)
	if err != nil {
		s.logger.Errorf("Could not list RDS Instances: %v", err)
		if len(awsResources) == 0 {
			return
		}
	}

	rdsInstances := lo.Map(awsResources, func(item awslib.AwsResource, _ int) rds.DBInstance {
		return item.(rds.DBInstance)
	})

	for _, item := range rdsInstances {
		assetChannel <- inventory.NewAssetEvent(
			inventory.AssetClassificationAwsRds,
			item.GetResourceArn(),
			item.GetResourceName(),
			inventory.WithRelatedAssetIds([]string{item.Identifier}),
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      item.GetRegion(),
				AccountID:   s.AccountId,
				AccountName: s.AccountName,
				ServiceName: "AWS RDS",
			}),
		)
	}
}
