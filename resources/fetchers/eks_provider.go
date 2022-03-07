package fetchers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

type EKSProvider struct {
	client *eks.Client
}

func NewEksProvider(cfg aws.Config) *EKSProvider {
	svc := eks.New(cfg)
	return &EKSProvider{
		client: svc,
	}
}

func (provider EKSProvider) DescribeCluster(ctx context.Context, clusterName string) (*eks.DescribeClusterResponse, error) {
	input := &eks.DescribeClusterInput{
		Name: &clusterName,
	}
	req := provider.client.DescribeClusterRequest(input)
	response, err := req.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster %s from eks , error - %w", clusterName, err)
	}

	return response, err
}
