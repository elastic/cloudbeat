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
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"gotest.tools/gotestsum/log"
)

type BucketDescription struct {
	Name         string
	SSEAlgorithm string
}

type S3 interface {
	DescribeBuckets(ctx context.Context) ([]awslib.AwsResource, error)
}

type Provider struct {
	log    *logp.Logger
	client *s3.Client
}

func NewProvider(cfg aws.Config, log *logp.Logger) *Provider {
	client := s3.NewFromConfig(cfg)
	return &Provider{
		log,
		client,
	}
}

func (p Provider) DescribeBuckets(ctx context.Context) ([]awslib.AwsResource, error) {
	clientBuckets, err := p.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		log.Errorf("Could not list s3 buckets: %v", err)
		return nil, err
	}

	var result []awslib.AwsResource

	for _, clientBucket := range clientBuckets.Buckets {
		sseAlgorithm := p.getBucketEncryptionAlgorithm(ctx, clientBucket.Name)

		result = append(result, BucketDescription{*clientBucket.Name, sseAlgorithm})
	}

	return result, nil
}

func (p Provider) getBucketEncryptionAlgorithm(ctx context.Context, bucketName *string) string {
	encryption, err := p.client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{Bucket: bucketName})

	if err != nil {
		p.log.Warnf("Could not get encryption for bucket %s. Error: %v", *bucketName, err)
	} else {
		if len(encryption.ServerSideEncryptionConfiguration.Rules) > 0 {
			return string(encryption.ServerSideEncryptionConfiguration.Rules[0].ApplyServerSideEncryptionByDefault.SSEAlgorithm)
		}
	}

	return ""
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
