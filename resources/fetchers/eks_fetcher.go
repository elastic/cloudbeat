package fetchers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

const EKSType = "aws-eks"

type EKSFetcher struct {
	cfg         EKSFetcherConfig
	eksProvider *EKSProvider
}

type EKSFetcherConfig struct {
	BaseFetcherConfig
	ClusterName string `config:"clusterName"`
}

type EKSResource struct {
	*eks.DescribeClusterResponse
}

func NewEKSFetcher(awsCfg aws.Config, cfg EKSFetcherConfig) (Fetcher, error) {
	eks := NewEksProvider(awsCfg)

	return &EKSFetcher{
		cfg:         cfg,
		eksProvider: eks,
	}, nil
}

func (f EKSFetcher) Fetch(ctx context.Context) ([]PolicyResource, error) {
	results := make([]PolicyResource, 0)

	result, err := f.eksProvider.DescribeCluster(ctx, f.cfg.ClusterName)
	results = append(results, EKSResource{result})

	return results, err
}

func (f EKSFetcher) Stop() {
}

//TODO: Add resource id logic to all AWS resources
func (r EKSResource) GetID() string {
	return ""
}

func (r EKSResource) GetData() interface{} {
	return r
}
