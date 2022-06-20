package add_cluster_id

import (
	"testing"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

type ClusterHelperTestSuite struct {
	suite.Suite

	log *logp.Logger
}

func TestClusterHelperTestSuite(t *testing.T) {
	s := new(ClusterHelperTestSuite)
	s.log = logp.NewLogger("cloudbeat_cluster_helper_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ClusterHelperTestSuite) TestClusterId() {
	kubeSystemNamespaceId := "123"
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-system",
			UID:  types.UID(kubeSystemNamespaceId),
		},
	}
	client := fake.NewSimpleClientset(ns)
	sut, err := newClusterHelper(client)
	s.NoError(err)

	s.Equal(kubeSystemNamespaceId, sut.ClusterId())
}

func (s *ClusterHelperTestSuite) TestClusterIdNotFound() {
	kubeSystemNamespaceId := "123"
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kube-sys",
			UID:  types.UID(kubeSystemNamespaceId),
		},
	}
	client := fake.NewSimpleClientset(ns)
	_, err := newClusterHelper(client)
	s.Error(err)
}
