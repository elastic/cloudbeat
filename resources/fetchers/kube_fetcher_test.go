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
	"context"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MockProvider(objects ...runtime.Object) func(s string, options kubernetes.KubeClientOptions) (k8s.Interface, error) {
	return func(s string, options kubernetes.KubeClientOptions) (k8s.Interface, error) {
		return k8sfake.NewSimpleClientset(objects...), nil
	}
}

func TestKubeFetcherFetchNoResources(t *testing.T) {
	kubeFetcher, err := (&KubeFactory{}).CreateFrom(KubeApiFetcherConfig{}, MockProvider())

	assert.NoError(t, err)

	results, err := kubeFetcher.Fetch(context.TODO())
	assert.Nil(t, err, "Fetcher was not able to fetch kubernetes resources")

	assert.Equal(t, 0, len(results))

	kubeFetcher.Stop()
}

func TestKubeFetcherFetchASinglePod(t *testing.T) {
	pod := v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "kube-system",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "nginx",
					Image:           "nginx",
					ImagePullPolicy: "Always",
				},
			},
		},
	}
	kubeFetcher, err := (&KubeFactory{}).CreateFrom(KubeApiFetcherConfig{}, MockProvider(&pod))

	assert.NoError(t, err)

	results, err := kubeFetcher.Fetch(context.TODO())
	assert.Nil(t, err, "Fetcher was not able to fetch kubernetes resources")

	assert.Equal(t, 1, len(results))
	assert.Equal(t, &pod, results[0].GetData())
}
