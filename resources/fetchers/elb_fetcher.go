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

type ELBDesc []elasticloadbalancing.LoadBalancerDescription
type ELBResource struct {
	ELBDesc
}

func NewELBFetcher(awsCfg aws.Config, cfg ELBFetcherConfig) (Fetcher, error) {
	elb := NewELBProvider(awsCfg)

	return &ELBFetcher{
		elbProvider: elb,
		cfg:         cfg,
	}, nil
}

func (f ELBFetcher) Fetch(ctx context.Context) ([]PolicyResource, error) {
	results := make([]PolicyResource, 0)

	result, err := f.elbProvider.DescribeLoadBalancer(ctx, f.cfg.LoadBalancerNames)
	results = append(results, ELBResource{result})

	return results, err
}

func (f ELBFetcher) Stop() {
}

//TODO: Add resource id logic to all AWS resources
func (res ELBResource) GetID() string {
	return ""
}

func (r ELBResource) GetData() interface{} {
	return r
}
