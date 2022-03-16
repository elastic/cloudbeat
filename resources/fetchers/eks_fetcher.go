package fetchers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/elastic/cloudbeat/resources/fetching"
)

const EKSType = "aws-eks"

type EKSFetcher struct {
	cfg         EKSFetcherConfig
	eksProvider *EKSProvider
}

type EKSFetcherConfig struct {
	fetching.BaseFetcherConfig
	ClusterName string `config:"clusterName"`
}

type EKSResource struct {
	*eks.DescribeClusterResponse
}

func NewEKSFetcher(awsCfg AwsFetcherConfig, cfg EKSFetcherConfig) (fetching.Fetcher, error) {
	eks := NewEksProvider(awsCfg.Config)

	return &EKSFetcher{
		cfg:         cfg,
		eksProvider: eks,
	}, nil
}

func (f EKSFetcher) Fetch(ctx context.Context) ([]fetching.Resource, error) {
	results := make([]fetching.Resource, 0)

	result, err := f.eksProvider.DescribeCluster(ctx, f.cfg.ClusterName)
	results = append(results, EKSResource{result})

	return results, err
}

func (f EKSFetcher) Stop() {
}

//TODO: Add resource id logic to all AWS resources
func (r EKSResource) GetID() (string, error) {
	return "", nil
}

func (r EKSResource) GetData() interface{} {
	return r
}
