package fetchers

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
)

const ELBType = "aws-elb"
const ELBRegexTemplate = "([\\w-]+)-\\d+\\.%s.elb.amazonaws.com"

type ELBFetcher struct {
	cfg             ELBFetcherConfig
	elbProvider     *ELBProvider
	kubeInitOnce    sync.Once
	kubeClient      k8s.Interface
	lbRegexMatchers []*regexp.Regexp
}

type ELBFetcherConfig struct {
	BaseFetcherConfig
	Kubeconfig string `config:"Kubeconfig"`
}

type LoadBalancersDescription []elasticloadbalancing.LoadBalancerDescription

type ELBResource struct {
	LoadBalancersDescription
}

func NewELBFetcher(awsCfg AwsFetcherConfig, cfg ELBFetcherConfig) (Fetcher, error) {
	elb := NewELBProvider(awsCfg.Config)

	loadBalancerRegex := fmt.Sprintf(ELBRegexTemplate, awsCfg.Config.Region)

	return &ELBFetcher{
		elbProvider:     elb,
		cfg:             cfg,
		lbRegexMatchers: []*regexp.Regexp{regexp.MustCompile(loadBalancerRegex)},
	}, nil
}

func (f *ELBFetcher) Fetch(ctx context.Context) ([]FetchedResource, error) {
	var err error
	f.kubeInitOnce.Do(func() {
		f.kubeClient, err = kubernetes.GetKubernetesClient(f.cfg.Kubeconfig, kubernetes.KubeClientOptions{})
	})
	if err != nil {
		// Reset watcherlock if the watchers could not be initiated.
		watcherlock = sync.Once{}
		return nil, fmt.Errorf("could not initate Kubernetes watchers: %w", err)
	}
	return f.getData(ctx)
}

func (f *ELBFetcher) getData(ctx context.Context) ([]FetchedResource, error) {
	results := make([]FetchedResource, 0)

	balancers, err := f.GetLoadBalancers()
	if err != nil {
		return nil, fmt.Errorf("failed to load balancers from Kubernetes %w", err)
	}
	result, err := f.elbProvider.DescribeLoadBalancer(ctx, balancers)
	if err != nil {
		return nil, fmt.Errorf("failed to load balancers from ELB %w", err)
	}
	results = append(results, ELBResource{result})

	return results, err
}

func (f *ELBFetcher) GetLoadBalancers() ([]string, error) {
	ctx := context.Background()
	services, err := f.kubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Kuberenetes services:  %w", err)
	}
	loadBalancers := make([]string, 0)
	for _, service := range services.Items {
		for _, ingress := range service.Status.LoadBalancer.Ingress {
			for _, matcher := range f.lbRegexMatchers {
				if matcher.MatchString(ingress.Hostname) {
					// Extract the repository name out of the image name
					lbName := matcher.FindStringSubmatch(ingress.Hostname)[1]
					loadBalancers = append(loadBalancers, lbName)
				}
			}
		}
	}
	return loadBalancers, nil
}

func (f *ELBFetcher) Stop() {
}

// GetID TODO: Add resource id logic to all AWS resources
func (r ELBResource) GetID() string {
	return ""
}

func (r ELBResource) GetData() interface{} {
	return r
}
