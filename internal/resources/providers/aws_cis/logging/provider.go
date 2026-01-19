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
	"strings"

	s3Client "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/cloudtrail"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
)

type EnrichedTrail struct {
	cloudtrail.TrailInfo
	BucketInfo TrailBucket `json:"bucket_info"`
}

type TrailBucket struct {
	Policy  s3.BucketPolicy              `json:"policy,omitempty"`
	Logging s3.Logging                   `json:"logging,omitempty"`
	ACL     *s3Client.GetBucketAclOutput `json:"acl,omitempty"`
}

// logBucketError logs an error with appropriate level based on the error message.
// If the error is nil, it returns early. If the error message contains "NoSuchBucket",
// it uses Warnf, otherwise Errorf.
func (p *Provider) logBucketError(err error, bucketName, operation string) {
	if err == nil {
		return
	}
	
	errMsg := err.Error()
	if strings.Contains(errMsg, "NoSuchBucket") {
		p.log.Warnf("Error getting bucket %s for bucket %s: %v", operation, bucketName, err)
	} else {
		p.log.Errorf("Error getting bucket %s for bucket %s: %v", operation, bucketName, err)
	}
}

func (p *Provider) DescribeTrails(ctx context.Context) ([]awslib.AwsResource, error) {
	trails, trailsErr := p.trailProvider.DescribeTrails(ctx)
	if trailsErr != nil {
		return nil, trailsErr
	}

	enrichedTrails := make([]awslib.AwsResource, 0, len(trails))
	for _, info := range trails {
		if info.Trail.S3BucketName == nil {
			continue
		}
		bucketPolicy, policyErr := p.s3Provider.GetBucketPolicy(ctx, info.Trail.S3BucketName, *info.Trail.HomeRegion)
		p.logBucketError(policyErr, *info.Trail.S3BucketName, "policy")

		aclGrants, aclErr := p.s3Provider.GetBucketACL(ctx, info.Trail.S3BucketName, *info.Trail.HomeRegion)
		p.logBucketError(aclErr, *info.Trail.S3BucketName, "ACL")

		bucketLogging, loggingErr := p.s3Provider.GetBucketLogging(ctx, info.Trail.S3BucketName, *info.Trail.HomeRegion)
		p.logBucketError(loggingErr, *info.Trail.S3BucketName, "logging")

		enrichedTrails = append(enrichedTrails, EnrichedTrail{
			TrailInfo: info,
			BucketInfo: TrailBucket{
				ACL:     aclGrants,
				Policy:  bucketPolicy,
				Logging: bucketLogging,
			},
		})
	}

	return enrichedTrails, nil
}

func (e EnrichedTrail) GetResourceArn() string {
	if e.Trail.TrailARN == nil {
		return ""
	}
	return *e.Trail.TrailARN
}

func (e EnrichedTrail) GetResourceName() string {
	if e.Trail.Name == nil {
		return ""
	}
	return *e.Trail.Name
}

func (e EnrichedTrail) GetResourceType() string {
	return fetching.TrailType
}

func (e EnrichedTrail) GetRegion() string {
	if e.Trail.HomeRegion == nil {
		return ""
	}
	return *e.Trail.HomeRegion
}
