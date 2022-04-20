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
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MockProvider(client *k8sfake.Clientset) func(s string, options kubernetes.KubeClientOptions) (k8s.Interface, error) {
	return func(s string, options kubernetes.KubeClientOptions) (k8s.Interface, error) {
		return client, nil
	}
}

func clean(fetcher fetching.Fetcher) func() {
	return func() {
		fetcher.Stop()
		watcherlock = sync.Once{}
	}
}

// TODO: use go generics and unify these functions
func RolesPtrMap(items []rbacv1.Role) []interface{} {
	vsm := make([]interface{}, len(items))
	for i, _ := range items {
		vsm[i] = &items[i]
	}
	return vsm
}

func PodsPtrMap(items []v1.Pod) []interface{} {
	vsm := make([]interface{}, len(items))
	for i, _ := range items {
		vsm[i] = &items[i]
	}
	return vsm
}

func Map(resources []fetching.Resource) []interface{} {
	vsm := make([]interface{}, len(resources))
	for i, v := range resources {
		vsm[i] = v.GetData()
	}
	return vsm
}

// TODO: convert all tests to a single table test and use add more resource types
func TestKubeFetcherFetchNoResources(t *testing.T) {
	client := k8sfake.NewSimpleClientset()
	provider := MockProvider(client)
	kubeFetcher, err := (&KubeFactory{}).CreateFrom(KubeApiFetcherConfig{}, provider)

	assert.NoError(t, err)

	results, err := kubeFetcher.Fetch(context.TODO())
	assert.Nil(t, err, "Fetcher was not able to fetch kubernetes resources")

	assert.Equal(t, 0, len(results))

	t.Cleanup(clean(kubeFetcher))
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
	client := k8sfake.NewSimpleClientset(&pod)
	provider := MockProvider(client)
	kubeFetcher, err := (&KubeFactory{}).CreateFrom(KubeApiFetcherConfig{}, provider)

	assert.NoError(t, err)

	results, err := kubeFetcher.Fetch(context.TODO())
	assert.Nil(t, err, "Fetcher was not able to fetch kubernetes resources")

	require.Equal(t, 1, len(results))
	require.ElementsMatch(t, PodsPtrMap([]v1.Pod{pod}), Map(results))

	t.Cleanup(clean(kubeFetcher))
}

func TestKubeFetcherFetchTwoPods(t *testing.T) {
	pods := v1.PodList{Items: []v1.Pod{
		{
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
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod2",
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
		},
	}}
	client := k8sfake.NewSimpleClientset(&pods)
	provider := MockProvider(client)
	kubeFetcher, err := (&KubeFactory{}).CreateFrom(KubeApiFetcherConfig{}, provider)

	assert.NoError(t, err)

	results, err := kubeFetcher.Fetch(context.TODO())
	assert.Nil(t, err, "Fetcher was not able to fetch kubernetes resources")

	require.ElementsMatch(t, PodsPtrMap(pods.Items), Map(results))

	t.Cleanup(clean(kubeFetcher))
}

func TestKubeFetcherFetchThreeRoles(t *testing.T) {
	roles := rbacv1.RoleList{Items: []rbacv1.Role{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-role1",
				Namespace: "default",
			},
			Rules: []rbacv1.PolicyRule{
				{Verbs: []string{"get"}},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-role2",
				Namespace: "default",
			},
			Rules: []rbacv1.PolicyRule{
				{Verbs: []string{"list"}},
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-role3",
				Namespace: "default",
			},
			Rules: []rbacv1.PolicyRule{
				{Verbs: []string{"create"}},
			},
		},
	}}
	client := k8sfake.NewSimpleClientset(&roles)
	provider := MockProvider(client)
	kubeFetcher, err := (&KubeFactory{}).CreateFrom(KubeApiFetcherConfig{}, provider)

	assert.NoError(t, err)

	results, err := kubeFetcher.Fetch(context.TODO())
	assert.Nil(t, err, "Fetcher was not able to fetch kubernetes resources")

	require.ElementsMatch(t, RolesPtrMap(roles.Items), Map(results))

	t.Cleanup(clean(kubeFetcher))
}
