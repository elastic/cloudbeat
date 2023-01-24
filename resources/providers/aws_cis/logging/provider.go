package logging

import (
	"context"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/providers/awslib/cloudtrail"
)

type EnrichedTrail struct {
	cloudtrail.TrailInfo
	grants []s3Types.Grant
}

func (p *Provider) DescribeTrails(ctx context.Context) ([]awslib.AwsResource, error) {
	trails, trailsErr := p.trailProvider.DescribeTrails(ctx)
	if trailsErr != nil {
		return nil, trailsErr
	}

	var enrichedTrails []awslib.AwsResource
	for _, trail := range trails {
		aclGrants, aclErr := p.s3Provider.GetBucketACL(ctx, &trail.BucketName, trail.Region)
		if aclErr != nil {
			aclGrants = []s3Types.Grant{}
			p.log.Errorf("Error getting bucket ACL for bucket %s: %v", trail.BucketName, aclErr)
		}

		enrichedTrails = append(enrichedTrails, EnrichedTrail{
			TrailInfo: trail,
			grants:    aclGrants,
		})
	}

	return enrichedTrails, nil
}

func (e EnrichedTrail) GetResourceArn() string {
	return e.TrailARN
}

func (e EnrichedTrail) GetResourceName() string {
	return e.Name
}

func (e EnrichedTrail) GetResourceType() string {
	return fetching.TrailType
}
