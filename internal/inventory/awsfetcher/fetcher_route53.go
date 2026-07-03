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
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/route53"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

type route53Fetcher struct {
	logger        *clog.Logger
	provider      route53Provider
	AccountId     string
	AccountName   string
	statusHandler statushandler.StatusHandlerAPI
}

type route53Provider interface {
	ListRecords(ctx context.Context) ([]awslib.AwsResource, error)
}

func newRoute53Fetcher(logger *clog.Logger, identity *cloud.Identity, provider route53Provider, statusHandler statushandler.StatusHandlerAPI) inventory.AssetFetcher {
	return &route53Fetcher{
		logger:        logger,
		provider:      provider,
		AccountId:     identity.Account,
		AccountName:   identity.AccountAlias,
		statusHandler: statusHandler,
	}
}

func (f *route53Fetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	f.logger.Info("Fetching Route53 records")
	defer f.logger.Info("Fetching Route53 records - Finished")

	resources, err := f.provider.ListRecords(ctx)
	if err != nil {
		f.logger.Errorf("Could not list Route53 records: %v", err)
		awslib.ReportMissingPermission(f.statusHandler, err)
	}

	for _, item := range resources {
		record, ok := item.(route53.Record)
		if !ok {
			continue
		}

		assetChannel <- inventory.NewAssetEvent(
			inventory.AssetClassificationAwsRoute53Record,
			item.GetResourceArn(),
			item.GetResourceName(),
			inventory.WithRawAsset(record),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      item.GetRegion(),
				AccountID:   f.AccountId,
				AccountName: f.AccountName,
				ServiceName: "AWS Route 53",
			}),
			inventory.WithEntityAttributes(buildRoute53Attributes(record)),
		)
	}
}

// buildRoute53Attributes maps a record's non-ECS fields into entity.attributes using
// UpperCamelCase keys. Empty values are omitted.
func buildRoute53Attributes(record route53.Record) map[string]any {
	attrs := map[string]any{}
	if record.Type != "" {
		attrs["Type"] = record.Type
	}
	if len(record.Records) > 0 {
		attrs["ResourceRecords"] = record.Records
	}
	if record.Weight != nil {
		attrs["Weight"] = *record.Weight
	}
	if record.RoutingRegion != "" {
		attrs["Region"] = record.RoutingRegion
	}
	if record.ZoneID != "" {
		attrs["ZoneID"] = record.ZoneID
	}
	if record.ZoneName != "" {
		attrs["ZoneName"] = record.ZoneName
	}
	if record.AliasTargetDNS != "" {
		attrs["AliasTargetDNS"] = record.AliasTargetDNS
	}
	if record.AliasTargetZoneID != "" {
		attrs["AliasTargetZoneId"] = record.AliasTargetZoneID
	}
	if record.HealthCheckID != "" {
		attrs["HealthCheckId"] = record.HealthCheckID
	}
	return attrs
}
