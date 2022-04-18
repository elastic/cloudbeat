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
