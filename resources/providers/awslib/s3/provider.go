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

package s3

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3Client "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"gotest.tools/gotestsum/log"
)

func NewProvider(cfg aws.Config, log *logp.Logger) *Provider {
	client := s3Client.NewFromConfig(cfg)
	return &Provider{
		log:    log,
		client: client,
		region: cfg.Region,
	}
}

func (p Provider) DescribeBuckets(ctx context.Context) ([]awslib.AwsResource, error) {
	clientBuckets, err := p.client.ListBuckets(ctx, &s3Client.ListBucketsInput{})
	if err != nil {
		log.Errorf("Could not list s3 buckets: %v", err)
		return nil, err
	}

	var result []awslib.AwsResource

	for _, clientBucket := range clientBuckets.Buckets {
		bucketRegion := p.getBucketRegion(ctx, clientBucket.Name)
		if bucketRegion == p.region {
			sseAlgorithm := p.getBucketEncryptionAlgorithm(ctx, clientBucket.Name)

			result = append(result, BucketDescription{*clientBucket.Name, sseAlgorithm})
		}
	}

	return result, nil
}

func (p Provider) getBucketEncryptionAlgorithm(ctx context.Context, bucketName *string) string {
	encryption, err := p.client.GetBucketEncryption(ctx, &s3Client.GetBucketEncryptionInput{Bucket: bucketName})

	if err != nil {
		shouldLogError := true
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "ServerSideEncryptionConfigurationNotFoundError" {
				shouldLogError = false
			}
		}

		if shouldLogError {
			p.log.Errorf("Could not get encryption for bucket %s. Error: %v", *bucketName, err)
		}

		return ""
	}

	if len(encryption.ServerSideEncryptionConfiguration.Rules) <= 0 {
		return ""
	}

	return string(encryption.ServerSideEncryptionConfiguration.Rules[0].ApplyServerSideEncryptionByDefault.SSEAlgorithm)
}

func (p Provider) getBucketRegion(ctx context.Context, bucketName *string) string {
	location, err := p.client.GetBucketLocation(ctx, &s3Client.GetBucketLocationInput{Bucket: bucketName})
	if err != nil {
		p.log.Errorf("Could not get bucket location for bucket %s. Error: %v", *bucketName, err)
		return ""
	}

	region := string(location.LocationConstraint)
	// Region us-east-1 have a LocationConstraint of null.
	if region == "" {
		region = "us-east-1"
	}

	return region
}

func (b BucketDescription) GetResourceArn() string {
	return fmt.Sprintf("arn:aws:s3:::%s", b.Name)
}

func (b BucketDescription) GetResourceName() string {
	return b.Name
}

func (b BucketDescription) GetResourceType() string {
	return fetching.S3Type
}
