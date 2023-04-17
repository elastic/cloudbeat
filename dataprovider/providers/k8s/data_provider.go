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
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/dataprovider/types"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/gofrs/uuid"
)

const (
	namespace        = "kube-system"
	clusterNameField = "orchestrator.cluster.name"
)

var uuidNamespace = uuid.Must(uuid.FromString("971a1103-6b5d-4b60-ab3d-8a339a58c6c8"))

type DataProvider struct {
	log       *logp.Logger
	cfg       *config.Config
	info      version.CloudbeatVersionInfo
	cluster   string
	clusterID string
	nodeID    string
}

func New(options ...Option) DataProvider {
	kdp := DataProvider{}
	for _, opt := range options {
		opt(&kdp)
	}
	return kdp
}

func (k DataProvider) FetchData(resource string, id string) (types.Data, error) {
	switch resource {
	case fetchers.ProcessResourceType, fetchers.FSResourceType:
		id = uuid.NewV5(uuidNamespace, k.clusterID+k.nodeID+id).String()
	case fetching.CloudContainerMgmt, fetching.CloudIdentity, fetching.CloudLoadBalancer, fetching.CloudContainerRegistry:
		id = uuid.NewV5(uuidNamespace, k.clusterID).String()
	}
	return types.Data{
		ResourceID:  id,
		VersionInfo: k.info,
	}, nil
}

func (k DataProvider) EnrichEvent(event *beat.Event, _ fetching.ResourceMetadata) error {
	name := k.cluster
	if name == "" {
		return nil
	}
	_, err := event.Fields.Put(clusterNameField, name)
	return err
}
