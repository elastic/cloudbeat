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

package conditions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestKubeLeaseIsLeader(t *testing.T) {
	holder := "my_cloudbeat"
	t.Setenv("POD_NAME", holder)

	leases := v1.LeaseList{Items: []v1.Lease{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Lease",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "elastic-agent-cluster-leader",
				Namespace: "kube-system",
			},
			Spec: v1.LeaseSpec{
				HolderIdentity: &holder,
			},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Lease",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-lease",
				Namespace: "kube-system",
			},
			Spec: v1.LeaseSpec{
				HolderIdentity: &holder,
			},
		},
	}}

	client := k8sfake.NewSimpleClientset(&leases)
	provider := NewLeaderLeaseProvider(context.TODO(), client)

	result, err := provider.IsLeader()
	assert.Nil(t, err, "Fetcher was not able to fetch kubernetes resources")
	assert.True(t, result)
}

func TestKubeLeaseIsNotLeader(t *testing.T) {
	holder := "other_cloudbeat"
	t.Setenv("POD_NAME", "my_cloudbeat")

	leases := v1.LeaseList{Items: []v1.Lease{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Lease",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "elastic-agent-cluster-leader",
				Namespace: "kube-system",
			},
			Spec: v1.LeaseSpec{
				HolderIdentity: &holder,
			},
		},
	}}

	client := k8sfake.NewSimpleClientset(&leases)
	provider := NewLeaderLeaseProvider(context.TODO(), client)

	result, err := provider.IsLeader()
	assert.Nil(t, err, "Fetcher was not able to fetch kubernetes resources")
	assert.False(t, result)
}

func TestKubeLeaseNoLeader(t *testing.T) {
	holder := "other_cloudbeat"
	t.Setenv("POD_NAME", "my_cloudbeat")

	leases := v1.LeaseList{Items: []v1.Lease{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Lease",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-lease",
				Namespace: "kube-system",
			},
			Spec: v1.LeaseSpec{
				HolderIdentity: &holder,
			},
		},
	}}

	client := k8sfake.NewSimpleClientset(&leases)
	provider := NewLeaderLeaseProvider(context.TODO(), client)

	result, err := provider.IsLeader()
	assert.Error(t, err)
	assert.False(t, result)
}
