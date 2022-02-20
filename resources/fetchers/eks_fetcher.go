package fetchers

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

const EKSType = "aws-eks"

type EKSResource struct {
	*eks.DescribeClusterResponse
}

type EKSFetcher struct {
	cfg         EKSFetcherConfig
	eksProvider *EKSProvider
}

type EKSFetcherConfig struct {
	BaseFetcherConfig
	ClusterName string `config:"clusterName"`
}

func NewEKSFetcher(awsCfg aws.Config, cfg EKSFetcherConfig) (Fetcher, error) {
	eks := NewEksProvider(awsCfg)

	return &EKSFetcher{
		cfg:         cfg,
		eksProvider: eks,
	}, nil
}

func (f *EKSFetcher) Fetch(ctx context.Context) ([]PolicyResource, error) {
	results := make([]PolicyResource, 0)

	result, err := f.eksProvider.DescribeCluster(ctx, f.cfg.ClusterName)
	results = append(results, EKSResource{DescribeClusterResponse: result})

	return results, err
}

func (f *EKSFetcher) Stop() {
}

// GetID TODO: Add resource id logic to all AWS resources
func (r EKSResource) GetID() string {
	return ""
}

func (r EKSResource) GetData() interface{} {
	return r
}