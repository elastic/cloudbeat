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
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/sns"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/statushandler"
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
			inventory.AssetClassificationAwsSnsTopic,
			"topic:arn:test-topic",
			"test-topic",
			inventory.WithRawAsset(awsResource),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS SNS",
			}),
		),
	}

	logger := testhelper.NewLogger(t)
	provider := newMockSnsProvider(t)
	provider.EXPECT().ListTopicsWithSubscriptions(mock.Anything).Return(in, nil)

	msh := statushandler.NewMockStatusHandlerAPI(t)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newSNSFetcher(logger, identity, provider, msh)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
