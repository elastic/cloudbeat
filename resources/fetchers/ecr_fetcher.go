package fetchers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

const ECRType = "aws-ecr"

type ECRFetcher struct {
	cfg         ECRFetcherConfig
	ecrProvider *ECRProvider
}

type ECRFetcherConfig struct {
	BaseFetcherConfig
}

type ecrRepos []ecr.Repository
type ECRResource struct {
	ecrRepos
}

func NewECRFetcher(awsCfg aws.Config, cfg ECRFetcherConfig) (Fetcher, error) {
	ecr := NewEcrProvider(awsCfg)

	return &ECRFetcher{
		cfg:         cfg,
		ecrProvider: ecr,
	}, nil
}

func (f ECRFetcher) Fetch(ctx context.Context) ([]PolicyResource, error) {
	results := make([]PolicyResource, 0)

	// TODO - The provider should get a list of the repositories it needs to check, and not check the entire ECR account`
	repositories, err := f.ecrProvider.DescribeAllECRRepositories(ctx)
	results = append(results, ECRResource{repositories})

	return results, err
}

func (f ECRFetcher) Stop() {}

//TODO: Add resource id logic to all AWS resources
func (res ECRResource) GetID() string {
	return ""
}
func (res ECRResource) GetData() interface{} {
	return res
}
