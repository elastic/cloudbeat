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

package fetchers

import (
	"encoding/gob"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

type KubeFactory struct {
}

func init() {
	manager.Factories.ListFetcherFactory(fetching.KubeAPIType, &KubeFactory{})
	gob.Register(K8sResource{})
	gob.Register(ECRResource{})
	gob.Register(ELBResource{})
	gob.Register(EKSResource{})
	gob.Register(IAMResource{})
	gob.Register(kubernetes.Pod{})
	gob.Register(kubernetes.Secret{})
	gob.Register(kubernetes.Role{})
	gob.Register(kubernetes.RoleBinding{})
	gob.Register(kubernetes.ClusterRole{})
	gob.Register(kubernetes.ClusterRoleBinding{})
	gob.Register(kubernetes.NetworkPolicy{})
	gob.Register(kubernetes.PodSecurityPolicy{})
}

func (f *KubeFactory) Create(c *common.Config) (fetching.Fetcher, error) {
	cfg := KubeApiFetcherConfig{}
	err := c.Unpack(&cfg)
	if err != nil {
		return nil, err
	}

	return f.CreateFrom(cfg)
}

func (f *KubeFactory) CreateFrom(cfg KubeApiFetcherConfig) (fetching.Fetcher, error) {
	fe := &KubeFetcher{
		cfg:      cfg,
		watchers: make([]kubernetes.Watcher, 0),
	}

	return fe, nil
}
