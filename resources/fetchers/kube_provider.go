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
	"fmt"
	"reflect"

	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sResource struct {
	Data interface{}
}

const k8sObjMetadataField = "ObjectMeta"

func GetKubeData(watchers []kubernetes.Watcher) []fetching.Resource {
	ret := make([]fetching.Resource, 0)

	for _, watcher := range watchers {
		rs := watcher.Store().List()

		for _, r := range rs {
			nullifyManagedFields(r)
			resource, ok := r.(kubernetes.Resource)

			if !ok {
				logp.L().Errorf("Bad resource: %#v does not implement kubernetes.Resource", r)
				continue
			}

			err := addTypeInformationToKubeResource(resource)
			if err != nil {
				logp.L().Errorf("Bad resource: %w", err)
				continue
			} // See https://github.com/kubernetes/kubernetes/issues/3030

			ret = append(ret, K8sResource{resource})
		}
	}

	return ret
}

func (r K8sResource) GetID() (string, error) {
	k8sObj := reflect.ValueOf(r.Data)
	metadata, ok := k8sObj.FieldByName(k8sObjMetadataField).Interface().(metav1.ObjectMeta)
	if !ok {
		return "", fmt.Errorf("failed to retrieve object metadata")
	}

	uid := metadata.UID
	return string(uid), nil
}

func (r K8sResource) GetData() interface{} {
	return r.Data
}

// nullifyManagedFields ManagedFields field contains fields with dot that prevent from elasticsearch to index
// the events.
func nullifyManagedFields(resource interface{}) {
	switch val := resource.(type) {
	case *kubernetes.Pod:
		val.ManagedFields = nil
	case *kubernetes.Secret:
		val.ManagedFields = nil
	case *kubernetes.Role:
		val.ManagedFields = nil
	case *kubernetes.RoleBinding:
		val.ManagedFields = nil
	case *kubernetes.ClusterRole:
		val.ManagedFields = nil
	case *kubernetes.ClusterRoleBinding:
		val.ManagedFields = nil
	case *kubernetes.PodSecurityPolicy:
		val.ManagedFields = nil
	case *kubernetes.NetworkPolicy:
		val.ManagedFields = nil
	}
}
