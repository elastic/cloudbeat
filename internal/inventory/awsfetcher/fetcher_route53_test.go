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
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/route53"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

func TestRoute53Fetcher_Fetch(t *testing.T) {
	record1 := route53.Record{
		Name:              "www.example.com.",
		Type:              "A",
		Records:           []string{"203.0.113.10"},
		Weight:            pointers.Ref(int64(10)),
		RoutingRegion:     "us-east-1",
		ZoneID:            "Z123",
		ZoneName:          "example.com.",
		AliasTargetDNS:    "alias.example.com.",
		AliasTargetZoneID: "Z456",
		HealthCheckID:     "hc-1",
	}

	record2 := route53.Record{
		Name:     "example.com.",
		Type:     "NS",
		ZoneID:   "Z123",
		ZoneName: "example.com.",
	}

	in := []awslib.AwsResource{record1, record2}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsRoute53Record,
			"arn:aws:route53:::hostedzone/Z123/recordset/www.example.com./A",
			"www.example.com.",
			inventory.WithRawAsset(record1),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      "global",
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS Route 53",
			}),
			inventory.WithEntityDetails(map[string]any{
				"Type":              "A",
				"ResourceRecords":   []string{"203.0.113.10"},
				"Weight":            int64(10),
				"Region":            "us-east-1",
				"ZoneID":            "Z123",
				"ZoneName":          "example.com.",
				"AliasTargetDNS":    "alias.example.com.",
				"AliasTargetZoneId": "Z456",
				"HealthCheckId":     "hc-1",
			}),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsRoute53Record,
			"arn:aws:route53:::hostedzone/Z123/recordset/example.com./NS",
			"example.com.",
			inventory.WithRawAsset(record2),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      "global",
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS Route 53",
			}),
			inventory.WithEntityDetails(map[string]any{
				"Type":     "NS",
				"ZoneID":   "Z123",
				"ZoneName": "example.com.",
			}),
		),
	}

	logger := testhelper.NewLogger(t)
	provider := newMockRoute53Provider(t)
	provider.EXPECT().ListRecords(mock.Anything).Return(in, nil)

	msh := statushandler.NewMockStatusHandlerAPI(t)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newRoute53Fetcher(logger, identity, provider, msh)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
