package fetchers

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AwsIdentity struct {
	Account *string
	Arn     *string
	UserId  *string
}

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
func (provider *AWSIdentityProvider) GetMyIdentity(ctx context.Context) (*AwsIdentity, error) {
	input := &sts.GetCallerIdentityInput{}
	request := provider.client.GetCallerIdentityRequest(input)
	response, err := request.Send(ctx)
	if err != nil {
		return nil, err
	}

	identity := &AwsIdentity{
		Account: response.Account,
		UserId:  response.UserId,
		Arn:     response.Arn,
	}
	return identity, nil
}
