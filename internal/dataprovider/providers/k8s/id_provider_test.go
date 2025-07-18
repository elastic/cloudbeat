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

	"github.com/stretchr/testify/assert"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	fetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/k8s"
)

func Test_k8sIdProvider_GetIdInCluster(t *testing.T) {
	tests := []struct {
		name     string
		want     string
		resource string
		id       string
	}{
		{
			name: "unknown resource should return the raw id",
			id:   "metadata_id",
			want: "metadata_id",
		},
		{
			name:     "CAAS resource should return cluster id",
			id:       "metadata_id",
			want:     "cluster_id",
			resource: fetching.CloudContainerMgmt,
		},
		{
			name:     "process reource should add cluster id and node id",
			id:       "metadata_id",
			want:     "cluster_idnode_idmetadata_id",
			resource: fetchers.ProcessResourceType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewIdProvider("cluster_id", "node_id")
			p, ok := provider.(*idProvider)
			assert.True(t, ok, "NewIdProvider should return *idProvider")
			data := p.getIdInCluster(tt.resource, tt.id)
			assert.Equal(t, tt.want, data)
		})
	}
}

func Test_k8sIdProvider_GetId(t *testing.T) {
	tests := []struct {
		name     string
		want     string
		resource string
		id       string
	}{
		{
			name: "unknown resource should return the raw id",
			id:   "metadata_id",
			want: "16d3e3ff-583a-5b0d-86e2-1f7c6675dfb3",
		},
		{
			name:     "CAAS resource should return cluster id",
			id:       "metadata_id",
			want:     "e0b7df83-548e-5f75-b460-23bf3a96f679",
			resource: fetching.CloudContainerMgmt,
		},
		{
			name:     "process reource should add cluster id and node id",
			id:       "metadata_id",
			want:     "76389fcc-375d-5acc-8e0a-db814362fc50",
			resource: fetchers.ProcessResourceType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewIdProvider("cluster_id", "node_id")
			data := p.GetId(tt.resource, tt.id)
			assert.Equal(t, tt.want, data)
		})
	}
}
