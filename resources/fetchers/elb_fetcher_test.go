package fetchers

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
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

const (
	elbRegex = "([\\w-]+)-\\d+\\.us1-east.elb.amazonaws.com"
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

	var tests = []struct {
		ns                  string
		loadBalancerIngress []v1.LoadBalancerIngress
		lbResponse          []elasticloadbalancing.LoadBalancerDescription
	}{
		{
			"test_namespace",
			[]v1.LoadBalancerIngress{
				{
					Hostname: "adda9cdc89b13452e92d48be46858d37-1423035038.us-east-2.elb.amazonaws.com",
				},
			},
			[]elasticloadbalancing.LoadBalancerDescription{{
				Instances: []elasticloadbalancing.Instance{},
			}},
		},
		{
			"test_namespace",
			[]v1.LoadBalancerIngress{},
			[]elasticloadbalancing.LoadBalancerDescription{},
		},
	}
	for _, test := range tests {
		//Need to add services
		kubeclient := k8sfake.NewSimpleClientset()

		services := &v1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "apps/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testing_pod",
				Namespace: test.ns,
			},
			Status: v1.ServiceStatus{
				LoadBalancer: v1.LoadBalancerStatus{
					Ingress: test.loadBalancerIngress,
				},
			},
			Spec: v1.ServiceSpec{},
		}
		_, err := kubeclient.CoreV1().Services(test.ns).Create(context.Background(), services, metav1.CreateOptions{})

		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		elbProvider := &awslib.MockedELBLoadBalancerDescriber{}
		elbProvider.EXPECT().DescribeLoadBalancer(mock.Anything, mock.Anything).Return(test.lbResponse, nil)

		regexMatchers := []*regexp.Regexp{regexp.MustCompile(elbRegex)}

		elbFetcher := ELBFetcher{
			cfg:             ELBFetcherConfig{},
			elbProvider:     elbProvider,
			kubeClient:      kubeclient,
			lbRegexMatchers: regexMatchers,
		}

		ctx := context.Background()

		expectedResource := ELBResource{test.lbResponse}
		result, err := elbFetcher.Fetch(ctx)
		s.Nil(err)
		s.Equal(1, len(result))

		elbResource := result[0].(ELBResource)
		s.Equal(expectedResource, elbResource)
	}
}

func (s *ElbFetcherTestSuite) TestCreateFetcherErrorCases() {

	var tests = []struct {
		ns                  string
		loadBalancerIngress []v1.LoadBalancerIngress
		error               error
	}{
		{
			"test_namespace",
			[]v1.LoadBalancerIngress{
				{
					Hostname: "adda9cdc89b13452e92d48be46858d37-1423035038.us-east-2.elb.amazonaws.com",
				},
			},
			fmt.Errorf("elb error")},
	}
	for _, test := range tests {
		//Need to add services
		kubeclient := k8sfake.NewSimpleClientset()

		services := &v1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "apps/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testing_pod",
				Namespace: test.ns,
			},
			Status: v1.ServiceStatus{
				LoadBalancer: v1.LoadBalancerStatus{
					Ingress: test.loadBalancerIngress,
				},
			},
			Spec: v1.ServiceSpec{},
		}
		_, err := kubeclient.CoreV1().Services(test.ns).Create(context.Background(), services, metav1.CreateOptions{})

		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		elbProvider := &awslib.MockedELBLoadBalancerDescriber{}
		elbProvider.EXPECT().DescribeLoadBalancer(mock.Anything, mock.Anything).Return(nil, test.error)

		regexMatchers := []*regexp.Regexp{regexp.MustCompile(elbRegex)}

		elbFetcher := ELBFetcher{
			cfg:             ELBFetcherConfig{},
			elbProvider:     elbProvider,
			kubeClient:      kubeclient,
			lbRegexMatchers: regexMatchers,
		}

		ctx := context.Background()

		result, err := elbFetcher.Fetch(ctx)
		s.Nil(result)
		s.NotNil(err)
		s.EqualError(err, fmt.Sprintf("failed to load balancers from ELB %s", test.error.Error()))
	}
}
