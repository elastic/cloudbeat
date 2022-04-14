package fetchers

import (
	"context"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"regexp"
	"testing"
)

type ElbFetcherTestSuite struct {
	suite.Suite
}

func TestElbFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(ElbFetcherTestSuite))
}

func (s *ElbFetcherTestSuite) SetupTest() {

}

func (s *ElbFetcherTestSuite) TestCreateFetcher() {

	//Need to add services
	kubeclient := k8sfake.NewSimpleClientset()

	ns := "test_namespace"

	containers := []v1.Container{
		{
			Image: "nginx:1.120",
			Name:  "nginx",
		},
	}

	pods := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "apps/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testing_pod",
			Namespace: ns,
		},
		Spec: kubernetes.PodSpec{
			NodeName:   "testnode",
			Containers: containers,
		},
	}
	_, err := kubeclient.CoreV1().Pods(ns).Create(context.Background(), pods, metav1.CreateOptions{})

	mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
	mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

	// Needs to use the same services
	elbProvider := &awslib.MockedELBLoadBalancerDescriber{}

	regex := "([\\w-]+)-\\d+\\.us1-east.elb.amazonaws.com"
	regexMatchers := []*regexp.Regexp{regexp.MustCompile(regex)}

	elbFetcher := ELBFetcher{
		cfg:             ELBFetcherConfig{},
		elbProvider:     elbProvider,
		kubeClient:      kubeclient,
		lbRegexMatchers: regexMatchers,
	}

	ctx := context.Background()

	expectedResource := ELBResource{}
	result, err := elbFetcher.Fetch(ctx)
	s.NotNil(err)

	elbResource := result[0].(ELBResource)

	s.Equal(expectedResource, elbResource)

}
