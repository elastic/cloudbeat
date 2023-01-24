// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
