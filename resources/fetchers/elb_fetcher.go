package fetchers

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/beats/v7/libbeat/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"regexp"
	"sync"

	"github.com/elastic/beats/v7/cloudbeat/resources"
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

func NewELBFetcher(awsConfig resources.AwsFetcherConfig, cfg ELBFetcherConfig) (resources.Fetcher, error) {
	elb := NewELBProvider(awsConfig.Config)

	awsConfig
	elbR := fmt.Sprintf(ELBRegexTemplate, awsConfig.Config.Region)

	return &ELBFetcher{
		elbProvider:     elb,
		cfg:             cfg,
		lbRegexMatchers: []*regexp.Regexp{regexp.MustCompile(elbR)},
	}, nil
}

func (f *ELBFetcher) Fetch(ctx context.Context) ([]resources.FetcherResult, error) {
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

func (f *ELBFetcher) getData(ctx context.Context) ([]resources.FetcherResult, error) {
	results := make([]resources.FetcherResult, 0)
	balancers, err := f.GetLoadBalancers()
	if err != nil {
		logp.Error(fmt.Errorf("failed to load balancers from Kubernetes %w", err))
		return nil, err
	}
	result, err := f.elbProvider.DescribeLoadBalancer(ctx, balancers)
	results = append(results, resources.FetcherResult{
		Type:     ELBType,
		Resource: result,
	})

	return results, err
}

func (f *ELBFetcher) GetLoadBalancers() ([]string, error) {
	ctx := context.Background()
	services, err := f.kubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
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
	if err != nil {
		logp.Error(fmt.Errorf("failed to get all services  - %w", err))
		return nil, err
	}

	return loadBalancers, nil
}

func (f *ELBFetcher) Stop() {
}
