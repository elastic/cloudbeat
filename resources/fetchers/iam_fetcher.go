package fetchers

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

const IAMType = "aws-iam"

type IAMFetcher struct {
	iamProvider *IAMProvider
	cfg         IAMFetcherConfig
}

type IAMFetcherConfig struct {
	BaseFetcherConfig
	RoleName string `config:"roleName"`
}

type IAMResource struct {
	Data interface{}
}

func NewIAMFetcher(awsCfg aws.Config, cfg IAMFetcherConfig) (Fetcher, error) {
	iam := NewIAMProvider(awsCfg)

	return &IAMFetcher{
		cfg:         cfg,
		iamProvider: iam,
	}, nil
}

func (f IAMFetcher) Fetch(ctx context.Context) ([]PolicyResource, error) {
	results := make([]PolicyResource, 0)

	result, err := f.iamProvider.GetIAMRolePermissions(ctx, f.cfg.RoleName)
	results = append(results, IAMResource{result})

	return results, err
}

func (f IAMFetcher) Stop() {
}

//TODO: Add resource id logic to all AWS resources
func (r IAMResource) GetID() string {
	return ""
}

func (r IAMResource) GetData() interface{} {
	return r.Data
}
