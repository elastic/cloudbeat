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
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func TestIAMPolicyFetcher_Fetch(t *testing.T) {
	now := time.Now()

	policy1 := iam.Policy{
		Policy: types.Policy{
			Arn:          pointers.Ref("arn:aws:iam::0000:policy/policy-1"),
			PolicyName:   pointers.Ref("policy-1"),
			PolicyId:     pointers.Ref("178263"),
			CreateDate:   &now,
			UpdateDate:   &now,
			Description:  pointers.Ref("test"),
			IsAttachable: true,
			Path:         pointers.Ref("/"),
			Tags: []types.Tag{
				{Key: pointers.Ref("key-1"), Value: pointers.Ref("value-1")},
				{Key: pointers.Ref("key-2"), Value: pointers.Ref("value-2")},
			},
		},
		Document: map[string]any{
			"Version": "2012-10-17",
			"Statement": []map[string]any{
				{
					"Effect":   "Allow",
					"Action":   []string{"read", "update", "delete"},
					"Resource": []string{"s3/bucket", "s3/bucket/*"},
				},
				{
					"Effect":   "Deny",
					"Action":   []string{"delete"},
					"Resource": []string{"s3/bucket"},
				},
			},
		},
		Roles: []types.PolicyRole{
			{RoleId: pointers.Ref("roleId-1"), RoleName: pointers.Ref("roleName-1")},
			{RoleId: pointers.Ref("roleId-2"), RoleName: pointers.Ref("roleName-2")},
		},
	}

	policy2 := iam.Policy{
		Policy: types.Policy{
			Arn:        pointers.Ref("arn:aws:iam::0000:policy/policy-2"),
			PolicyName: pointers.Ref("policy-2"),
			Tags: []types.Tag{
				{Key: pointers.Ref("key-1"), Value: pointers.Ref("value-1")},
			},
		},
		Document: map[string]any{
			"Version": "2012-10-17",
			"Statement": map[string]any{
				"Effect":   "Allow",
				"Action":   "read",
				"Resource": "*",
			},
		},
		Roles: []types.PolicyRole{
			{RoleId: pointers.Ref("roleId-1"), RoleName: pointers.Ref("roleName-1")},
		},
	}

	policy3 := iam.Policy{
		Policy: types.Policy{
			Arn:        pointers.Ref("arn:aws:iam::0000:policy/policy-3"),
			PolicyName: pointers.Ref("policy-3"),
		},
	}

	in := []awslib.AwsResource{policy1, nil, policy2, policy3}

	cloudField := inventory.Cloud{
		Provider:    inventory.AwsCloudProvider,
		Region:      "global",
		AccountID:   "123",
		AccountName: "alias",
		ServiceName: "AWS IAM",
	}

	expected := []inventory.AssetEvent{
		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsIamPolicy,
			"arn:aws:iam::0000:policy/policy-1",
			"policy-1",
			inventory.WithRawAsset(policy1),
			inventory.WithCloud(cloudField),
			inventory.WithLabels(map[string]string{
				"key-1": "value-1",
				"key-2": "value-2",
			}),
		),

		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsIamPolicy,
			"arn:aws:iam::0000:policy/policy-2",
			"policy-2",
			inventory.WithRawAsset(policy2),
			inventory.WithCloud(cloudField),
			inventory.WithLabels(map[string]string{
				"key-1": "value-1",
			}),
		),

		inventory.NewAssetEvent(
			inventory.AssetClassificationAwsIamPolicy,
			"arn:aws:iam::0000:policy/policy-3",
			"policy-3",
			inventory.WithRawAsset(policy3),
			inventory.WithCloud(cloudField),
		),
	}

	logger := logp.NewLogger("test_fetcher_iam_role")
	provider := newMockIamPolicyProvider(t)
	provider.EXPECT().GetPolicies(mock.Anything).Return(in, nil)

	identity := &cloud.Identity{Account: "123", AccountAlias: "alias"}
	fetcher := newIamPolicyFetcher(logger, identity, provider)

	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}
