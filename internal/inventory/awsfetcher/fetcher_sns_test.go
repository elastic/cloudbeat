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

	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/sns"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestSNSFetcher_Fetch(t *testing.T) {
	awsResource := sns.TopicInfo{
		Topic: types.Topic{
			TopicArn: pointers.Ref("topic:arn:test-topic"),
		},
		Subscriptions: []types.Subscription{},
	}

	in := []awslib.AwsResource{awsResource}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassification{
				Category:    inventory.CategoryInfrastructure,
				SubCategory: inventory.SubCategoryIntegration,
				Type:        inventory.TypeNotificationService,
				SubType:     inventory.SubTypeSNSTopic,
			},
			"topic:arn:test-topic",
			"test-topic",
			inventory.WithRawAsset(awsResource),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Account: inventory.AssetCloudAccount{
					Id:   "123",
					Name: "alias",
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS SNS",
				},
			}),
		),
	}

	logger := logp.NewLogger("test_fetcher_sns_instance")
	provider := newMockSnsProvider(t)
	provider.EXPECT().ListTopicsWithSubscriptions(mock.Anything).Return(in, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newSNSFetcher(logger, identity, provider)

	collectResourcesAndMatch(t, fetcher, expected)
}
