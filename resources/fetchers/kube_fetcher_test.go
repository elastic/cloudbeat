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
	"fmt"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"reflect"
	"sync"
	"testing"
)

type KubeFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

func TestKubeFetcherTestSuite(t *testing.T) {
	s := new(KubeFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_kube_fetcher_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *KubeFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *KubeFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func MockProvider(client *k8sfake.Clientset) KubeClientProvider {
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

func MapItems(resources runtime.Object) []interface{} {
	r := reflect.ValueOf(resources)
	f := reflect.Indirect(r).FieldByName("Items")
	items := f.Interface()
	// Finding a way to avoid this switch case could be nice
	switch items := items.(type) {
	case []v1.Pod:
		return PtrMap(items)
	case []rbacv1.Role:
		return PtrMap(items)
	default:
		return nil
	}
}

func PtrMap[In any](items []In) []interface{} {
	vsm := make([]interface{}, len(items))
	for i := range items {
		vsm[i] = &items[i]
	}
	return vsm
}

func Map[In fetching.Resource](resources []In) []interface{} {
	vsm := make([]interface{}, len(resources))
	for i, v := range resources {
		vsm[i] = v.GetData()
	}
	return vsm
}

func (s *KubeFetcherTestSuite) TestKubeFetcher_TestFetch() {
	myPod := v1.Pod{
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
	threePods := v1.PodList{Items: []v1.Pod{
		myPod,
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
	threeRoles := rbacv1.RoleList{Items: []rbacv1.Role{
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
	tests := []runtime.Object{
		&v1.PodList{},
		&v1.PodList{Items: []v1.Pod{myPod}},
		&threePods,
		&threeRoles,
	}

	for i, tt := range tests {
		s.Run(fmt.Sprintf("Kube api test %v", i), func() {
			client := k8sfake.NewSimpleClientset(tt)
			provider := MockProvider(client)

			kubeFetcher, err := (&KubeFactory{}).CreateFrom(s.log, KubeApiFetcherConfig{}, s.resourceCh, provider)

			s.NoError(err)

			err = kubeFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
			results := testhelper.CollectResources(s.resourceCh)

			s.Nil(err, "Fetcher was not able to fetch resources from kube api")
			require.ElementsMatch(s.T(), MapItems(tests[i]), Map(results))

			s.T().Cleanup(clean(kubeFetcher))
		})
	}
}
