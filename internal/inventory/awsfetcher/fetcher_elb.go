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
)

type elbFetcher struct {
	logger      *logp.Logger
	v1          v1Provider
	v2          v2Provider
	AccountId   string
	AccountName string
}

type v1Provider interface {
	DescribeAllLoadBalancers(context.Context) ([]awslib.AwsResource, error)
}
type v2Provider interface {
	DescribeLoadBalancers(context.Context) ([]awslib.AwsResource, error)
}

func newElbFetcher(logger *logp.Logger, identity *cloud.Identity, v1Provider v1Provider, v2Provider v2Provider) inventory.AssetFetcher {
	return &elbFetcher{
		logger:      logger,
		v1:          v1Provider,
		v2:          v2Provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

type elbDescribeFunc func(context.Context) ([]awslib.AwsResource, error)

func (f *elbFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	resourcesToFetch := []struct {
		name           string
		function       elbDescribeFunc
		classification inventory.AssetClassification
	}{
		{"Elastic Load Balancers v1", f.v1.DescribeAllLoadBalancers, inventory.AssetClassificationAwsElbV1},
		{"Elastic Load Balancers v2", f.v2.DescribeLoadBalancers, inventory.AssetClassificationAwsElbV2},
	}
	for _, r := range resourcesToFetch {
		f.fetch(ctx, r.name, r.function, r.classification, assetChannel)
	}
}

func (f *elbFetcher) fetch(ctx context.Context, resourceName string, function elbDescribeFunc, classification inventory.AssetClassification, assetChannel chan<- inventory.AssetEvent) {
	f.logger.Infof("Fetching %s", resourceName)
	defer f.logger.Infof("Fetching %s - Finished", resourceName)

	awsResources, err := function(ctx)
	if err != nil {
		f.logger.Errorf("Could not fetch %s: %v", resourceName, err)
		return
	}

	for _, item := range awsResources {
		assetChannel <- inventory.NewAssetEvent(
			classification,
			item.GetResourceArn(),
			item.GetResourceName(),
			inventory.WithRawAsset(item),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      item.GetRegion(),
				AccountID:   f.AccountId,
				AccountName: f.AccountName,
				ServiceName: "AWS Networking",
			}),
		)
	}
}
