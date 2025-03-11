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
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
)

type lambdaFetcher struct {
	logger      *clog.Logger
	provider    lambdaProvider
	AccountId   string
	AccountName string
}

type (
	lambdaDescribeFunc func(context.Context) ([]awslib.AwsResource, error)
	lambdaProvider     interface {
		ListAliases(context.Context, string, string) ([]awslib.AwsResource, error)
		ListEventSourceMappings(context.Context) ([]awslib.AwsResource, error)
		ListFunctions(context.Context) ([]awslib.AwsResource, error)
		ListLayers(context.Context) ([]awslib.AwsResource, error)
	}
)

func newLambdaFetcher(logger *clog.Logger, identity *cloud.Identity, provider lambdaProvider) inventory.AssetFetcher {
	return &lambdaFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (s *lambdaFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	resourcesToFetch := []struct {
		name           string
		function       lambdaDescribeFunc
		classification inventory.AssetClassification
	}{
		{"Lambda Event Source Mappings", s.provider.ListEventSourceMappings, inventory.AssetClassificationAwsLambdaEventSourceMapping},
		{"Lambda Functions", s.provider.ListFunctions, inventory.AssetClassificationAwsLambdaFunction},
		{"Lambda Layers", s.provider.ListLayers, inventory.AssetClassificationAwsLambdaLayer},
	}
	for _, r := range resourcesToFetch {
		s.fetch(ctx, r.name, r.function, r.classification, assetChannel)
	}
}

func (s *lambdaFetcher) fetch(ctx context.Context, resourceName string, function lambdaDescribeFunc, classification inventory.AssetClassification, assetChannel chan<- inventory.AssetEvent) {
	s.logger.Infof("Fetching %s", resourceName)
	defer s.logger.Infof("Fetching %s - Finished", resourceName)

	awsResources, err := function(ctx)
	if err != nil {
		s.logger.Errorf("Could not fetch %s: %v", resourceName, err)
		return
	}

	for _, item := range awsResources {
		var id string = item.GetResourceArn()
		if id == "" { // e.g. LambdaEventSourceMappings
			id = item.GetResourceName()
		}
		assetChannel <- inventory.NewAssetEvent(
			classification,
			[]string{id},
			item.GetResourceName(),
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      item.GetRegion(),
				AccountID:   s.AccountId,
				AccountName: s.AccountName,
				ServiceName: "AWS Lambda",
			}),
		)
	}
}
