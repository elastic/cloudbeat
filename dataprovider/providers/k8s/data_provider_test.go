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

package k8s

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider/types"
	"github.com/elastic/cloudbeat/resources/fetchers/process"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/version"

	// "github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/stretchr/testify/assert"
)

var (
	logger      = logp.NewLogger("cloudbeat_k8s_common_data_provider_test")
	versionInfo = version.CloudbeatVersionInfo{
		Version: version.CloudbeatVersion(),
		Policy:  version.PolicyVersion(),
		Kubernetes: version.Version{
			Version: ".",
		},
	}
)

var (
	clusterID = "kube-system_id"
	nodeID    = "node_id"
)

func Test_k8sDataProvider_FetchData(t *testing.T) {
	tests := []struct {
		name     string
		options  []Option
		want     types.Data
		resource string
	}{
		{
			name: "should return the metadata.id when the resource is unknown",
			want: types.Data{
				ResourceID: "metadata_id",
				VersionInfo: version.CloudbeatVersionInfo{
					Version: version.CloudbeatVersion(),
					Policy:  version.PolicyVersion(),
					Kubernetes: version.Version{
						Version: ".",
					},
				},
			},
			options: []Option{
				WithLogger(logger),
				WithVersionInfo(versionInfo),
			},
		},
		{
			name: "should add cluster id",
			want: types.Data{
				ResourceID: "d3069a00-f692-57c3-9094-9741c52526ff",
				VersionInfo: version.CloudbeatVersionInfo{
					Version: version.CloudbeatVersion(),
					Policy:  version.PolicyVersion(),
					Kubernetes: version.Version{
						Version: ".",
					},
				},
			},
			resource: fetching.CloudContainerMgmt,
			options: []Option{
				WithLogger(logger),
				WithVersionInfo(versionInfo),
				WithClusterID(clusterID),
			},
		},
		{
			name: "should add cluster and node id",
			want: types.Data{
				ResourceID: "0afa24c0-4069-5b7d-93cd-d334469e42c0",
				VersionInfo: version.CloudbeatVersionInfo{
					Version: version.CloudbeatVersion(),
					Policy:  version.PolicyVersion(),
					Kubernetes: version.Version{
						Version: ".",
					},
				},
			},
			resource: process.ProcessResourceType,
			options: []Option{
				WithLogger(logger),
				WithVersionInfo(versionInfo),
				WithClusterID(clusterID),
				WithNodeID(nodeID),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.options...)
			data, err := p.FetchData(tt.resource, "metadata_id")

			assert.NoError(t, err)
			assert.Equal(t, tt.want, data)
		})
	}
}

func TestK8sDataProvider_EnrichEvent(t *testing.T) {
	options := []Option{
		WithClusterName("test_cluster"),
		WithLogger(logger),
		WithConfig(&config.Config{
			Benchmark: config.CIS_K8S,
		}),
	}

	k := New(options...)
	e := &beat.Event{
		Fields: mapstr.M{},
	}
	err := k.EnrichEvent(e)
	assert.NoError(t, err)
	v, err := e.Fields.GetValue(clusterNameField)
	assert.NoError(t, err)
	assert.Equal(t, "test_cluster", v)
}
