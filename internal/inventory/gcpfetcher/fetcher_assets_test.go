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

package gcpfetcher

import (
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/inventory/testutil"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	gcpinventory "github.com/elastic/cloudbeat/internal/resources/providers/gcplib/inventory"
)

func TestAccountFetcher_Fetch_Assets(t *testing.T) {
	logger := clog.NewLogger("gcpfetcher_test")
	createAsset := func(assetType string) *gcpinventory.ExtendedGcpAsset {
		return &gcpinventory.ExtendedGcpAsset{
			Asset: &assetpb.Asset{
				Name:      "/projects/<project UUID>/some_resource",
				AssetType: assetType,
			},
			CloudAccount: &fetching.CloudAccountMetadata{
				AccountId:        "<project UUID>",
				AccountName:      "<project name>",
				OrganisationId:   "<org UUID>",
				OrganizationName: "<org name>",
			},
		}
	}
	expected := lo.Map(ResourcesToFetch, func(r ResourcesClassification, _ int) inventory.AssetEvent {
		return inventory.NewAssetEvent(
			r.classification,
			"/projects/<project UUID>/some_resource",
			"/projects/<project UUID>/some_resource",
			inventory.WithRawAsset(createAsset(r.assetType)),
			inventory.WithRelatedAssetIds([]string{}),
			inventory.WithCloud(inventory.Cloud{
				Provider:    inventory.GcpCloudProvider,
				AccountID:   "<project UUID>",
				AccountName: "<project name>",
				ProjectID:   "<org UUID>",
				ProjectName: "<org name>",
				ServiceName: r.assetType,
			}),
			inventory.WithLabels(map[string]string{}),
			inventory.WithTags([]string{}),
		)
	})

	provider := newMockInventoryProvider(t)
	for _, resource := range ResourcesToFetch {
		provider.EXPECT().mock.On("ListAssetTypes", mock.Anything, []string{resource.assetType}, mock.Anything).Run(func(args mock.Arguments) {
			ch := args.Get(2).(chan<- *gcpinventory.ExtendedGcpAsset)
			ch <- createAsset(resource.assetType)
			close(ch)
		})
	}
	fetcher := newAssetsInventoryFetcher(logger, provider)
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}

func TestAccountFetcher_EnrichAsset(t *testing.T) {
	var data = map[string]struct {
		resource    *assetpb.Resource    // input of GCP asset resource data
		enrichments inventory.AssetEvent // output of inventory asset ECS fields
	}{
		gcpinventory.IamRoleAssetType:   {},
		gcpinventory.CrmFolderAssetType: {},
		gcpinventory.ComputeNetworkAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"name": "network1",
				}),
			},
			enrichments: inventory.AssetEvent{
				Network: &inventory.Network{
					Name: "network1",
				},
			},
		},
		gcpinventory.CrmProjectAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"labels": map[string]any{"org": "security"},
				}),
			},
			enrichments: inventory.AssetEvent{
				Labels: map[string]string{"org": "security"},
			},
		},
		gcpinventory.StorageBucketAssetType:        {},
		gcpinventory.IamServiceAccountKeyAssetType: {},
		gcpinventory.CrmOrgAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"displayName": "org",
				}),
			},
			enrichments: inventory.AssetEvent{
				Organization: &inventory.Organization{
					Name: "org",
				},
			},
		},
		gcpinventory.ComputeInstanceAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"id":          "id",
					"name":        "name",
					"machineType": "machineType",
					"zone":        "zone",
					"labels":      map[string]any{"key": "value"},
				}),
			},
			enrichments: inventory.AssetEvent{
				Cloud: &inventory.Cloud{
					InstanceID:       "id",
					InstanceName:     "name",
					MachineType:      "machineType",
					AvailabilityZone: "zone",
				},
				Host: &inventory.Host{
					ID: "id",
				},
				Labels: map[string]string{"key": "value"},
			},
		},
		gcpinventory.ComputeFirewallAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"direction": "INGRESS",
					"name":      "default-allow-ssh",
				}),
			},
			enrichments: inventory.AssetEvent{
				Network: &inventory.Network{
					Direction: "INGRESS",
					Name:      "default-allow-ssh",
				},
			},
		},
		gcpinventory.ComputeSubnetworkAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"name":      "subnetwork",
					"stackType": "IPV4_ONLY",
				}),
			},
			enrichments: inventory.AssetEvent{
				Network: &inventory.Network{
					Name: "subnetwork",
					Type: "ipv4_only",
				},
			},
		},
		gcpinventory.IamServiceAccountAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"displayName": "service-account",
					"email":       "service-account@<project UUID>.iam.gserviceaccount.com",
				}),
			},
			enrichments: inventory.AssetEvent{
				User: &inventory.User{
					Name:  "service-account",
					Email: "service-account@<project UUID>.iam.gserviceaccount.com",
				},
			},
		},
		gcpinventory.GkeClusterAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"name": "cluster",
					"id":   "cluster-id",
				}),
			},
			enrichments: inventory.AssetEvent{
				Orchestrator: &inventory.Orchestrator{
					Type:        "kubernetes",
					ClusterName: "cluster",
					ClusterID:   "cluster-id",
				},
			},
		},
		gcpinventory.ComputeForwardingRuleAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"region": "region1",
				}),
			},
			enrichments: inventory.AssetEvent{
				Cloud: &inventory.Cloud{
					Region: "region1",
				},
			},
		},
		gcpinventory.CloudFunctionAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"name": "cloud-function",
					"url":  "https://cloud-function.com",
					"serviceConfig": map[string]any{
						"revision": "1",
					},
				}),
			},
			enrichments: inventory.AssetEvent{
				Fass: &inventory.Fass{
					Name:    "cloud-function",
					Version: "1",
				},
				URL: &inventory.URL{
					Full: "https://cloud-function.com",
				},
			},
		},
		gcpinventory.CloudRunService: {},
	}

	for _, r := range ResourcesToFetch {
		item, ok := data[r.assetType]
		if !ok {
			t.Errorf("Missing case for %s", r.assetType)
		}

		gcpAsset := &gcpinventory.ExtendedGcpAsset{
			Asset: &assetpb.Asset{
				Name:      "/projects/<project UUID>/some_resource",
				AssetType: r.assetType,
				Resource:  item.resource,
			},
			CloudAccount: &fetching.CloudAccountMetadata{
				AccountId:        "<project UUID>",
				AccountName:      "<project name>",
				OrganisationId:   "<org UUID>",
				OrganizationName: "<org name>",
			},
		}
		actual := getAssetEvent(r.classification, gcpAsset)
		expected := item.enrichments

		// Set the common fields that are not set in the enrichments
		expected.Event = actual.Event
		expected.Entity = actual.Entity
		expected.RawAttributes = actual.RawAttributes

		// Cloud is the only field where we have both common and enriched fields
		if expected.Cloud == nil {
			// Use the actual cloud fields when there are no cloud enrichments
			expected.Cloud = actual.Cloud
		} else {
			// Use common cloud fields when there are cloud enrichments
			expected.Cloud.Provider = actual.Cloud.Provider
			expected.Cloud.AccountID = actual.Cloud.AccountID
			expected.Cloud.AccountName = actual.Cloud.AccountName
			expected.Cloud.ProjectID = actual.Cloud.ProjectID
			expected.Cloud.ProjectName = actual.Cloud.ProjectName
			expected.Cloud.ServiceName = actual.Cloud.ServiceName
		}

		assert.Equalf(t, expected, actual, "%v failed", "EnrichAsset")
	}
}

func NewStructMap(data map[string]any) *structpb.Struct {
	dataStruct, err := structpb.NewStruct(data)
	if err != nil {
		panic(err)
	}
	return dataStruct
}
