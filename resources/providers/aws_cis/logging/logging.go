package logging

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/resources/providers/awslib/s3"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Client interface {
	DescribeTrails(ctx context.Context) ([]awslib.AwsResource, error)
}

type Provider struct {
	log           *logp.Logger
	s3Provider    *s3.Provider
	trailProvider *cloudtrail.Provider
}

func NewProvider(
	log *logp.Logger,
	cfg aws.Config,
	multiRegionTrailFactory awslib.CrossRegionFactory[cloudtrail.Client],
	multiRegionS3Factory awslib.CrossRegionFactory[s3.Client],
) *Provider {

	return &Provider{
		log:           log,
		s3Provider:    s3.NewProvider(cfg, log, multiRegionS3Factory),
		trailProvider: cloudtrail.NewProvider(cfg, log, multiRegionTrailFactory),
	}
}
