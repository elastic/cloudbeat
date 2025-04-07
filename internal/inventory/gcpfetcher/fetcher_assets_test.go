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
	"cloud.google.com/go/iam/apiv1/iampb"
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
		provider.EXPECT().mock.On("ListAllAssetTypesByName", mock.Anything, []string{resource.assetType}).Return([]*gcpinventory.ExtendedGcpAsset{createAsset(resource.assetType)}, nil)
	}
	fetcher := newAssetsInventoryFetcher(logger, provider)
	testutil.CollectResourcesAndMatch(t, fetcher, expected)
}

func TestAccountFetcher_EnrichAsset(t *testing.T) {
	logger := clog.NewLogger("gcpfetcher_test")
	defaultCloud := &inventory.Cloud{
		Provider:    inventory.GcpCloudProvider,
		AccountID:   "<project UUID>",
		AccountName: "<project name>",
		ProjectID:   "<org UUID>",
		ProjectName: "<org name>",
	}

	var data = map[string]struct {
		resource  *assetpb.Resource    // input of GCP asset resource data
		iamPolicy *iampb.Policy        // input of GCP asset iam policy data
		event     inventory.AssetEvent // output of inventory asset ECS fields
	}{
		gcpinventory.ComputeNetworkAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"name": "network1",
				}),
			},
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
				Network: &inventory.Network{
					Name: "network1",
				},
			},
		},
		gcpinventory.IamRoleAssetType: {
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
			},
		},
		gcpinventory.CrmFolderAssetType: {
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
			},
		},
		gcpinventory.CrmProjectAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"labels": map[string]any{"org": "security"},
					"tags":   map[string]any{"items": []any{"tag1", "tag2"}},
				}),
			},
			iamPolicy: &iampb.Policy{
				Bindings: []*iampb.Binding{
					{
						Role:    "roles/owner",
						Members: []string{"user:a", "user:b"},
					},
				},
			},
			event: inventory.AssetEvent{
				Cloud:  defaultCloud,
				Labels: map[string]string{"org": "security"},
				Tags:   []string{"tag1", "tag2"},
				Related: &inventory.Related{
					Entity: []string{"roles/owner", "user:a", "user:b"},
				},
			},
		},
		gcpinventory.StorageBucketAssetType: {
			iamPolicy: &iampb.Policy{
				Bindings: []*iampb.Binding{
					{
						Role:    "roles/owner",
						Members: []string{"user:a", "user:b"},
					},
				},
			},
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
				Related: &inventory.Related{
					Entity: []string{"roles/owner", "user:a", "user:b"},
				},
			},
		},
		gcpinventory.IamServiceAccountKeyAssetType: {
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
			},
		},
		gcpinventory.CrmOrgAssetType: {
			resource: &assetpb.Resource{
				Parent: "organizations/<org UUID>",
				Data: NewStructMap(map[string]any{
					"displayName": "org",
					"tags":        map[string]any{"items": []any{}},
				}),
			},
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
				Organization: &inventory.Organization{
					Name: "org",
				},
				Related: &inventory.Related{
					Entity: []string{"organizations/<org UUID>"},
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
					"networkInterfaces": []any{
						map[string]any{"name": "nic0", "network": "n1", "subnetwork": "s1"},
						map[string]any{"name": "nic1", "network": "n2", "subnetwork": "s2"},
					},
					"serviceAccounts": []any{
						map[string]any{"email": "sa1@<project UUID>.iam.gserviceaccount.com"},
						map[string]any{"email": "sa2@<project UUID>.iam.gserviceaccount.com"},
					},
					"disks": []any{
						map[string]any{"source": "disk1"},
						map[string]any{"source": "disk2"},
					},
				}),
			},
			event: inventory.AssetEvent{
				Cloud: &inventory.Cloud{
					Provider:         defaultCloud.Provider,
					AccountID:        defaultCloud.AccountID,
					AccountName:      defaultCloud.AccountName,
					ProjectID:        defaultCloud.ProjectID,
					ProjectName:      defaultCloud.ProjectName,
					InstanceID:       "id",
					InstanceName:     "name",
					MachineType:      "machineType",
					AvailabilityZone: "zone",
				},
				Host: &inventory.Host{
					ID: "id",
				},
				Labels: map[string]string{"key": "value"},
				Related: &inventory.Related{
					Entity: []string{"n1", "s1", "n2", "s2", "sa1@<project UUID>.iam.gserviceaccount.com", "sa2@<project UUID>.iam.gserviceaccount.com", "disk1", "disk2", "machineType", "zone"},
				},
			},
		},
		gcpinventory.ComputeFirewallAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"direction": "INGRESS",
					"name":      "default-allow-ssh",
					"network":   "default",
				}),
			},
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
				Network: &inventory.Network{
					Direction: "INGRESS",
					Name:      "default-allow-ssh",
				},
				Related: &inventory.Related{
					Entity: []string{"default"},
				},
			},
		},
		gcpinventory.ComputeSubnetworkAssetType: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{
					"name":      "subnetwork",
					"stackType": "IPV4_ONLY",
					"network":   "network",
				}),
			},
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
				Network: &inventory.Network{
					Name: "subnetwork",
					Type: "ipv4_only",
				},
				Related: &inventory.Related{
					Entity: []string{"network"},
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
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
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
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
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
			event: inventory.AssetEvent{
				Cloud: &inventory.Cloud{
					Provider:    defaultCloud.Provider,
					AccountID:   defaultCloud.AccountID,
					AccountName: defaultCloud.AccountName,
					ProjectID:   defaultCloud.ProjectID,
					ProjectName: defaultCloud.ProjectName,
					Region:      "region1",
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
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
				Fass: &inventory.Fass{
					Name:    "cloud-function",
					Version: "1",
				},
				URL: &inventory.URL{
					Full: "https://cloud-function.com",
				},
			},
		},
		gcpinventory.CloudRunService: {
			resource: &assetpb.Resource{
				Data: NewStructMap(map[string]any{}),
			},
			event: inventory.AssetEvent{
				Cloud: defaultCloud,
			},
		},
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
				IamPolicy: item.iamPolicy,
			},
			CloudAccount: &fetching.CloudAccountMetadata{
				AccountId:        defaultCloud.AccountID,
				AccountName:      defaultCloud.AccountName,
				OrganisationId:   defaultCloud.ProjectID,
				OrganizationName: defaultCloud.ProjectName,
			},
		}
		actual := getAssetEvent(*logger, r.classification, gcpAsset)
		expected := item.event
		expected.Cloud.ServiceName = r.assetType
		inventory.WithRawAsset(gcpAsset)(&expected)

		assert.Equalf(t, expected.RawAttributes, actual.RawAttributes, "Asset %v failed %v fields", r.assetType, "RawAttributes")
		assert.Equalf(t, expected.Related, actual.Related, "Asset %v failed %v fields", r.assetType, "Related")
		assert.Equalf(t, expected.Cloud, actual.Cloud, "Asset %v failed %v fields", r.assetType, "Cloud")
		assert.Equalf(t, expected.Network, actual.Network, "Asset %v failed %v fields", r.assetType, "Network")
		assert.Equalf(t, expected.URL, actual.URL, "Asset %v failed %v fields", r.assetType, "URL")
		assert.Equalf(t, expected.Host, actual.Host, "Asset %v failed %v fields", r.assetType, "Host")
		assert.Equalf(t, expected.User, actual.User, "Asset %v failed %v fields", r.assetType, "User")
		assert.Equalf(t, expected.Organization, actual.Organization, "Asset %v failed %v fields", r.assetType, "Organization")
		assert.Equalf(t, expected.Labels, actual.Labels, "Asset %v failed %v fields", r.assetType, "Labels")
		assert.ElementsMatch(t, expected.Tags, actual.Tags)
	}
}

func NewStructMap(data map[string]any) *structpb.Struct {
	dataStruct, err := structpb.NewValue(data)
	if err != nil {
		panic(err)
	}
	return dataStruct.GetStructValue()
}
