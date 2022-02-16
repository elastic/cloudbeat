package resources

import "github.com/aws/aws-sdk-go-v2/aws"

type AwsFetcherConfig struct {
	Config    aws.Config
	AccountID *string
}
