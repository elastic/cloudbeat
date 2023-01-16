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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3Client "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
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
		p.log.Errorf("Could not list s3 buckets: %v", err)
		return nil, err
	}

	var result []awslib.AwsResource

	for _, clientBucket := range clientBuckets.Buckets {
		bucketRegion, regionErr := p.getBucketRegion(ctx, clientBucket.Name)
		// If we could not get the region for a bucket, additional API calls for resources will probably fail, we should
		//	not describe this bucket.
		if regionErr != nil {
			p.log.Errorf("Could not get bucket location for bucket %s. Not describing this bucket. Error: %v", *clientBucket.Name, regionErr)
			continue
		}

		// If the bucket is not in the configured region additional API calls for resources will fail, we should not
		//  describe this bucket.
		if bucketRegion != p.region {
			p.log.Debugf("Bucket %s is in region %s and not in the configured region %s. Not describing this bucket", *clientBucket.Name, bucketRegion, p.region)
			continue
		}

		// Getting the bucket encryption, policy and versioning is not critical for the rest of the flow, so we should
		//	keep describing the bucket even if getting these objects fails.
		sseAlgorithm, encryptionErr := p.getBucketEncryptionAlgorithm(ctx, clientBucket.Name)
		if encryptionErr != nil {
			p.log.Errorf("Could not get encryption for bucket %s. Error: %v", *clientBucket.Name, encryptionErr)
		}

		bucketPolicy, policyErr := p.getBucketPolicy(ctx, clientBucket.Name)
		if policyErr != nil {
			p.log.Errorf("Could not get bucket policy for bucket %s. Error: %v", *clientBucket.Name, policyErr)
		}

		bucketVersioning, versioningErr := p.getBucketVersioning(ctx, clientBucket.Name)
		if versioningErr != nil {
			p.log.Errorf("Could not get bucket versioning for bucket %s. Err: %v", *clientBucket.Name, versioningErr)
		}

		result = append(result, BucketDescription{*clientBucket.Name, sseAlgorithm, bucketPolicy, bucketVersioning})
	}

	return result, nil
}

func (p Provider) getBucketEncryptionAlgorithm(ctx context.Context, bucketName *string) (string, error) {
	encryption, err := p.client.GetBucketEncryption(ctx, &s3Client.GetBucketEncryptionInput{Bucket: bucketName})

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "ServerSideEncryptionConfigurationNotFoundError" {
				p.log.Debugf("Bucket encryption for bucket %s does not exist", *bucketName)
				return "", nil
			}
		}

		return "", err
	}

	if len(encryption.ServerSideEncryptionConfiguration.Rules) <= 0 {
		return "", nil
	}

	return string(encryption.ServerSideEncryptionConfiguration.Rules[0].ApplyServerSideEncryptionByDefault.SSEAlgorithm), nil
}

func (p Provider) getBucketRegion(ctx context.Context, bucketName *string) (string, error) {
	location, err := p.client.GetBucketLocation(ctx, &s3Client.GetBucketLocationInput{Bucket: bucketName})
	if err != nil {
		return "", err
	}

	region := string(location.LocationConstraint)
	// Region us-east-1 have a LocationConstraint of null.
	if region == "" {
		region = "us-east-1"
	}

	return region, nil
}

func (p Provider) getBucketPolicy(ctx context.Context, bucketName *string) (BucketPolicy, error) {
	rawPolicy, err := p.client.GetBucketPolicy(ctx, &s3Client.GetBucketPolicyInput{Bucket: bucketName})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchBucketPolicy" {
				p.log.Debugf("Bucket policy for bucket %s does not exist", *bucketName)
				return nil, nil
			}
		}

		return nil, err
	}

	var bucketPolicy BucketPolicy
	jsonErr := json.Unmarshal([]byte(*rawPolicy.Policy), &bucketPolicy)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return bucketPolicy, nil
}

func (p Provider) getBucketVersioning(ctx context.Context, bucketName *string) (BucketVersioning, error) {
	bucketVersioning := BucketVersioning{false, false}
	bucketVersioningResponse, err := p.client.GetBucketVersioning(ctx, &s3Client.GetBucketVersioningInput{Bucket: bucketName})
	if err != nil {
		return bucketVersioning, err
	}

	if bucketVersioningResponse.Status == types.BucketVersioningStatusEnabled {
		bucketVersioning.Enabled = true
	}

	if bucketVersioningResponse.MFADelete == types.MFADeleteStatusEnabled {
		bucketVersioning.MfaDelete = true
	}

	return bucketVersioning, nil
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
