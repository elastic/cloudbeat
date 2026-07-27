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
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/eks"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

type eksFetcher struct {
	logger        *clog.Logger
	provider      eksProvider
	accountId     string
	accountName   string
	statusHandler statushandler.StatusHandlerAPI
}

type eksProvider interface {
	DescribeClusters(ctx context.Context) ([]awslib.AwsResource, error)
}

func newEKSFetcher(logger *clog.Logger, identity *cloud.Identity, provider eksProvider, statusHandler statushandler.StatusHandlerAPI) inventory.AssetFetcher {
	return &eksFetcher{
		logger:        logger,
		provider:      provider,
		accountId:     identity.Account,
		accountName:   identity.AccountAlias,
		statusHandler: statusHandler,
	}
}

func (f *eksFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	f.logger.Info("Fetching EKS Clusters")
	defer f.logger.Info("Fetching EKS Clusters - Finished")

	resources, err := f.provider.DescribeClusters(ctx)
	if err != nil {
		f.logger.Errorf("Could not list EKS clusters: %v", err)
		awslib.ReportMissingPermission(f.statusHandler, err)
	}

	for _, item := range resources {
		cluster, ok := item.(eks.Cluster)
		if !ok {
			continue
		}

		assetChannel <- inventory.NewAssetEvent(
			inventory.AssetClassificationAwsEksCluster,
			item.GetResourceArn(),
			item.GetResourceName(),
			inventory.WithRawAsset(cluster),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      item.GetRegion(),
				AccountID:   f.accountId,
				AccountName: f.accountName,
				ServiceName: "AWS EKS",
			}),
			inventory.WithEntityDetails(buildEKSDetails(cluster)),
			inventory.WithCreatedAt(cluster.CreatedAt),
		)
	}
}

// buildEKSDetails maps a cluster's non-ECS fields into entity.Details using
// UpperCamelCase keys. The endpoint-access booleans are always included as they are
// meaningful even when false; other empty values are omitted.
func buildEKSDetails(cluster eks.Cluster) map[string]any {
	details := map[string]any{
		"EndpointPublicAccess":  cluster.EndpointPublicAccess,
		"EndpointPrivateAccess": cluster.EndpointPrivateAccess,
	}
	if cluster.Status != "" {
		details["Status"] = cluster.Status
	}
	if cluster.Version != "" {
		details["Version"] = cluster.Version
	}
	if cluster.Endpoint != "" {
		details["Endpoint"] = cluster.Endpoint
	}
	if cluster.RoleArn != "" {
		details["RoleArn"] = cluster.RoleArn
	}
	if cluster.PlatformVersion != "" {
		details["PlatformVersion"] = cluster.PlatformVersion
	}
	if v := cluster.GetOwnerTag(); v != "" {
		details["OwnerTag"] = v
	}
	return details
}
