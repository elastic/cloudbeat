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

package dataprovider

import (
	"context"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/version"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/gofrs/uuid"
)

const (
	namespace = "kube-system"
)

var uuidNamespace = uuid.Must(uuid.FromString("971a1103-6b5d-4b60-ab3d-8a339a58c6c8"))

func NewCommonDataProvider(log *logp.Logger, cfg *config.Config) CommonDataProvider {
	return CommonDataProvider{
		log:             log,
		cfg:             cfg,
		k8sDataProvider: NewK8sDataProvider(log, cfg),
	}
}

// FetchCommonData fetches cluster id, node id and version info
func (c CommonDataProvider) FetchCommonData(ctx context.Context) (CommonDataInterface, error) {
	cm := CommonData{}
	cm.versionInfo.Version = version.CloudbeatVersion()
	cm.versionInfo.Policy = version.PolicyVersion()

	k8sCommonData := c.k8sDataProvider.CollectK8sData(ctx)
	cm.versionInfo.Kubernetes = k8sCommonData.serverVersion
	cm.clusterId = k8sCommonData.clusterId
	cm.nodeId = k8sCommonData.nodeId

	return cm, nil
}

func (cd CommonData) GetResourceId(metadata fetching.ResourceMetadata) string {
	switch metadata.Type {
	case fetchers.ProcessResourceType, fetchers.FSResourceType:
		return uuid.NewV5(uuidNamespace, cd.clusterId+cd.nodeId+metadata.ID).String()
	case fetching.CloudContainerMgmt, fetching.CloudIdentity, fetching.CloudLoadBalancer, fetching.CloudContainerRegistry:
		return uuid.NewV5(uuidNamespace, cd.clusterId+metadata.ID).String()
	default:
		return metadata.ID
	}
}

func (cd CommonData) GetData() CommonData {
	return cd
}

func (cd CommonData) GetVersionInfo() version.CloudbeatVersionInfo {
	return cd.versionInfo
}
