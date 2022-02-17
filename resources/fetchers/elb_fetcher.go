package fetchers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
)

const ELBType = "aws-elb"

type ELBFetcher struct {
	cfg         ELBFetcherConfig
	elbProvider *ELBProvider
}

type ELBFetcherConfig struct {
	BaseFetcherConfig
	LoadBalancerNames []string `config:"loadBalancers"`
}

type LoadBalancerDescription []elasticloadbalancing.LoadBalancerDescription

type ELBResource struct {
	LoadBalancerDescription
}

func NewELBFetcher(awsCfg aws.Config, cfg ELBFetcherConfig) (Fetcher, error) {
	elb := NewELBProvider(awsCfg)

	return &ELBFetcher{
		elbProvider: elb,
		cfg:         cfg,
	}, nil
}

func (f ELBFetcher) Fetch(ctx context.Context) ([]FetchedResource, error) {
	results := make([]FetchedResource, 0)

	result, err := f.elbProvider.DescribeLoadBalancer(ctx, f.cfg.LoadBalancerNames)
	results = append(results, ELBResource{result})

	return results, err
}

func (f ELBFetcher) Stop() {
}

//TODO: Add resource id logic to all AWS resources
func (r ELBResource) GetID() string {
	return ""
}

func (r ELBResource) GetData() interface{} {
	return r
}
