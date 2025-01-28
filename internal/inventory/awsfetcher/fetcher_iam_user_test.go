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

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestIAMUserFetcher_Fetch(t *testing.T) {
	now := time.Now()

	user1 := iam.User{
		Name:                "user-1",
		Arn:                 "arn:aws:iam::000:user/user-1",
		UserId:              "u-123123",
		LastAccess:          "2023-03-28T12:27:26+00:00",
		PasswordEnabled:     true,
		MfaActive:           true,
		PasswordLastChanged: "2023-03-28T12:27:26+00:00",
		AccessKeys: []iam.AccessKey{
			{
				Active:       true,
				HasUsed:      true,
				LastAccess:   "2023-03-28T12:27:26+00:00",
				RotationDate: "2023-03-28T12:27:26+00:00",
			},
		},
		MFADevices: []iam.AuthDevice{
			{
				IsVirtual: true,
				MFADevice: types.MFADevice{
					EnableDate:   &now,
					SerialNumber: pointers.Ref("123"),
					UserName:     pointers.Ref("user-1"),
				},
			},
		},
		InlinePolicies: []iam.PolicyDocument{
			{
				PolicyName: "inline-policy",
				Policy:     "policy",
			},
		},
		AttachedPolicies: []types.AttachedPolicy{
			{
				PolicyArn:  pointers.Ref("arn:aws:iam:1321312:policy/att-policy"),
				PolicyName: pointers.Ref("att-policy"),
			},
		},
	}

	user2 := iam.User{
		Name:       "user-2",
		Arn:        "arn:aws:iam::000:user/user-2",
		LastAccess: "2023-03-28T12:27:26+00:00",
	}

	in := []awslib.AwsResource{user1, user2}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsIamUser,
			[]string{"arn:aws:iam::000:user/user-1", "u-123123"},
			"user-1",
			inventory.WithRawAsset(user1),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   "global",
				Account: inventory.AssetCloudAccount{
					Id:   "123",
					Name: "alias",
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS IAM",
				},
			}),
		),

		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsIamUser,
			[]string{"arn:aws:iam::000:user/user-2"},
			"user-2",
			inventory.WithRawAsset(user2),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   "global",
				Account: inventory.AssetCloudAccount{
					Id:   "123",
					Name: "alias",
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS IAM",
				},
			}),
		),
	}

	logger := clog.NewLogger("test_fetcher_iam_user")
	provider := newMockIamUserProvider(t)
	provider.EXPECT().GetUsers(mock.Anything).Return(in, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newIamUserFetcher(logger, identity, provider)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
