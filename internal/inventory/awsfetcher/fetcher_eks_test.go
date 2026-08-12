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
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/eks"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/statushandler"
)

func TestEKSFetcher_Fetch(t *testing.T) {
	createdAt := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)

	cluster1 := eks.Cluster{
		Name:                  "prod-cluster",
		Arn:                   "arn:aws:eks:us-east-1:123:cluster/prod-cluster",
		Status:                "ACTIVE",
		Version:               "1.29",
		Endpoint:              "https://abc.eks.amazonaws.com",
		RoleArn:               "arn:aws:iam::123:role/eks-role",
		PlatformVersion:       "eks.5",
		EndpointPublicAccess:  true,
		EndpointPrivateAccess: false,
		Tags:                  map[string]string{"Owner": "team-infra"},
		CreatedAt:             &createdAt,
	}

	cluster2 := eks.Cluster{
		Name: "minimal-cluster",
		Arn:  "arn:aws:eks:us-east-1:123:cluster/minimal-cluster",
	}

	in := []awslib.AwsResource{cluster1, cluster2}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsEksCluster,
			"arn:aws:eks:us-east-1:123:cluster/prod-cluster",
			"prod-cluster",
			inventory.WithRawAsset(cluster1),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS EKS",
			}),
			inventory.WithEntityDetails(map[string]any{
				"EndpointPublicAccess":  true,
				"EndpointPrivateAccess": false,
				"Status":                "ACTIVE",
				"Version":               "1.29",
				"Endpoint":              "https://abc.eks.amazonaws.com",
				"RoleArn":               "arn:aws:iam::123:role/eks-role",
				"PlatformVersion":       "eks.5",
				"OwnerTag":              "team-infra",
			}),
			inventory.WithCreatedAt(&createdAt),
		),
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsEksCluster,
			"arn:aws:eks:us-east-1:123:cluster/minimal-cluster",
			"minimal-cluster",
			inventory.WithRawAsset(cluster2),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS EKS",
			}),
			inventory.WithEntityDetails(map[string]any{
				"EndpointPublicAccess":  false,
				"EndpointPrivateAccess": false,
			}),
			inventory.WithCreatedAt(nil),
		),
	}

	logger := testhelper.NewLogger(t)
	provider := newMockEksProvider(t)
	provider.EXPECT().DescribeClusters(mock.Anything).Return(in, nil)

	msh := statushandler.NewMockStatusHandlerAPI(t)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newEKSFetcher(logger, identity, provider, msh)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
