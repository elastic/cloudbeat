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

type ECRFetcherTestSuite struct {
	suite.Suite
}

func TestECRFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(ECRFetcherTestSuite))
}

func (s *ECRFetcherTestSuite) SetupTest() {

}

func (s *ECRFetcherTestSuite) TestCreateFetcher() {

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
	ecrProvider := &awslib.MockedEcrRepositoryDescriber{}

	//Maybe will need to change this texts
	regexTexts := []string{
		"^my-account\\.dkr\\.ecr\\.us1-east\\.amazonaws\\.com\\/([-\\w]+)[:,@]?",
		"public\\.ecr\\.aws\\/\\w+\\/([\\w-]+)\\:?",
	}
	regexMatchers := []*regexp.Regexp{
		regexp.MustCompile(regexTexts[0]),
		regexp.MustCompile(regexTexts[1]),
	}

	ecrFetcher := ECRFetcher{
		cfg:               ECRFetcherConfig{},
		ecrProvider:       ecrProvider,
		kubeClient:        kubeclient,
		repoRegexMatchers: regexMatchers,
	}

	ctx := context.Background()

	expectedResource := ECRResource{}
	result, err := ecrFetcher.Fetch(ctx)
	s.NotNil(err)

	elbResource := result[0].(ELBResource)

	s.Equal(expectedResource, elbResource)

}
