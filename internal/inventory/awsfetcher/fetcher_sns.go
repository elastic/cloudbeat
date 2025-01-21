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

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

type snsFetcher struct {
	logger      *clog.Logger
	provider    snsProvider
	AccountId   string
	AccountName string
}

type snsProvider interface {
	ListTopicsWithSubscriptions(ctx context.Context) ([]awslib.AwsResource, error)
}

func newSNSFetcher(logger *clog.Logger, identity *cloud.Identity, provider snsProvider) inventory.AssetFetcher {
	return &snsFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (s *snsFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	s.logger.Info("Fetching SNS Topics")
	defer s.logger.Info("Fetching SNS Topics - Finished")

	awsResources, err := s.provider.ListTopicsWithSubscriptions(ctx)
	if err != nil {
		s.logger.Errorf("Could not fetch SNS Topics: %v", err)
		return
	}

	for _, item := range awsResources {
		assetChannel <- inventory.NewAssetEvent(
			inventory.AssetClassificationAwsSnsTopic,
			item.GetResourceArn(),
			item.GetResourceName(),
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      item.GetRegion(),
				AccountID:   s.AccountId,
				AccountName: s.AccountName,
				ServiceName: "AWS SNS",
			}),
		)
	}
}
