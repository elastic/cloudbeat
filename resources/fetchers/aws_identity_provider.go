package fetchers

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWSIdentityProvider struct {
	client *sts.Client
}

func NewAWSIdentityProvider(cfg aws.Config) *AWSIdentityProvider {
	svc := sts.New(cfg)
	return &AWSIdentityProvider{
		client: svc,
	}
}

// GetMyIdentity This method will return your identity (Arn, user-id...)
func (provider AWSIdentityProvider) GetMyIdentity(ctx context.Context) (*sts.GetCallerIdentityResponse, error) {
	input := &sts.GetCallerIdentityInput{}
	request := provider.client.GetCallerIdentityRequest(input)
	response, err := request.Send(ctx)

	return response, err
}
