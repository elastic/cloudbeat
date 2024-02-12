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
	"github.com/gofrs/uuid"

	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	fetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/k8s"
)

var uuidNamespace = uuid.Must(uuid.FromString("971a1103-6b5d-4b60-ab3d-8a339a58c6c8"))

type idProvider struct {
	namespace uuid.UUID
	clusterID string
	nodeID    string
}

func NewIdProvider(clusterID, nodeID string) dataprovider.IdProvider {
	return &idProvider{
		namespace: uuidNamespace,
		clusterID: clusterID,
		nodeID:    nodeID,
	}
}

func (p *idProvider) GetId(resourceType string, resourceId string) string {
	idInCluster := p.getIdInCluster(resourceType, resourceId)
	return uuid.NewV5(p.namespace, idInCluster).String()
}

func (p *idProvider) getIdInCluster(resourceType string, resourceId string) string {
	id := resourceId
	switch resourceType {
	case fetchers.ProcessResourceType, fetchers.FSResourceType:
		id = p.clusterID + p.nodeID + resourceId
	case fetching.CloudContainerMgmt, fetching.CloudIdentity, fetching.CloudLoadBalancer, fetching.CloudContainerRegistry:
		id = p.clusterID
	}

	return id
}
