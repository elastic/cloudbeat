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
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestIAMRoleFetcher_Fetch(t *testing.T) {
	now := time.Now()

	role1 := iam.Role{
		Role: types.Role{
			RoleName:                 pointers.Ref("role-name-1"),
			Arn:                      pointers.Ref("arn:aws:iam::0000:role/role-name-1"),
			RoleLastUsed:             nil,
			Tags:                     nil,
			CreateDate:               &now,
			MaxSessionDuration:       pointers.Ref(int32(3600)),
			PermissionsBoundary:      nil,
			AssumeRolePolicyDocument: pointers.Ref("document"),
			Description:              pointers.Ref("EKS managed node group IAM role"),
			Path:                     pointers.Ref("/"),
			RoleId:                   pointers.Ref("17823618723"),
		},
	}

	role2 := iam.Role{
		Role: types.Role{
			RoleName:                 pointers.Ref("role-name-2"),
			Arn:                      pointers.Ref("arn:aws:iam::0000:role/role-name-2"),
			RoleLastUsed:             nil,
			Tags:                     nil,
			CreateDate:               &now,
			MaxSessionDuration:       pointers.Ref(int32(3600)),
			PermissionsBoundary:      nil,
			AssumeRolePolicyDocument: pointers.Ref("document"),
			Description:              pointers.Ref("EKS managed node group IAM role"),
			Path:                     pointers.Ref("/"),
			RoleId:                   pointers.Ref("17823618723"),
		},
	}

	in := []*iam.Role{&role1, nil, &role2}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsIamRole,
			"arn:aws:iam::0000:role/role-name-1",
			"role-name-1",
			inventory.WithRawAsset(role1),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      "global",
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS IAM",
			}),
			inventory.WithUser(inventory.User{
				ID:   "arn:aws:iam::0000:role/role-name-1",
				Name: "role-name-1",
			}),
		),

		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsIamRole,
			"arn:aws:iam::0000:role/role-name-2",
			"role-name-2",
			inventory.WithRawAsset(role2),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.AwsCloudProvider,
				Region:      "global",
				AccountID:   "123",
				AccountName: "alias",
				ServiceName: "AWS IAM",
			}),
			inventory.WithUser(inventory.User{
				ID:   "arn:aws:iam::0000:role/role-name-2",
				Name: "role-name-2",
			}),
		),
	}

	logger := logp.NewLogger("test_fetcher_iam_role")
	provider := newMockIamRoleProvider(t)
	provider.EXPECT().ListRoles(mock.Anything).Return(in, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newIamRoleFetcher(logger, identity, provider)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
