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

package cloudfront

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

func NewProvider(log *logp.Logger, cfg aws.Config) *Provider {
	return &Provider{
		log:    log,
		client: cloudfront.NewFromConfig(cfg),
	}
}

func (p Provider) DescribeDistributions(ctx context.Context) ([]awslib.AwsResource, error) {
	p.log.Debug("CloudFrontProvider.DescribeDistributions")
	var sdkDistributions []types.DistributionSummary
	params := &cloudfront.ListDistributionsInput{}

	for {
		distributions, err := p.client.ListDistributions(ctx, params)
		if err != nil {
			return nil, err
		}
		sdkDistributions = append(sdkDistributions, distributions.DistributionList.Items...)
		isTruncated := distributions.DistributionList.IsTruncated
		if isTruncated == nil || !(pointers.Deref(isTruncated)) {
			break
		}
		params.Marker = distributions.DistributionList.NextMarker
	}

	result := make([]awslib.AwsResource, len(sdkDistributions))
	for _, distribution := range sdkDistributions {
		result = append(result, Distribution{
			DistributionSummary: distribution,
			awsAccount:          accountFromARN(distribution.ARN),
		})
	}

	p.log.Debug("CloudFrontProvider.DescribeDistributions got %d items", len(result))
	return result, nil
}

func (d Distribution) GetRegion() string {
	return ""
}

func (d Distribution) GetResourceArn() string {
	return pointers.Deref(d.ARN)
}

func (d Distribution) GetResourceName() string {
	elements := strings.Split(pointers.Deref(d.ARN), "/")
	return elements[len(elements)-1]
}

func (d Distribution) GetResourceType() string {
	return fetching.CloudFrontDistributionType
}

func (p Provider) DescribeKeyValueStores(ctx context.Context) ([]awslib.AwsResource, error) {
	p.log.Debug("CloudFrontProvider.DescribeKeyValueStores")
	var sdkKeyValueStores []types.KeyValueStore
	params := &cloudfront.ListKeyValueStoresInput{}

	for {
		kvs, err := p.client.ListKeyValueStores(ctx, params)
		if err != nil {
			return nil, err
		}
		sdkKeyValueStores = append(sdkKeyValueStores, kvs.KeyValueStoreList.Items...)
		nextMarker := kvs.KeyValueStoreList.NextMarker
		if nextMarker == nil || pointers.Deref(nextMarker) == "" {
			break
		}
		params.Marker = nextMarker
	}

	result := make([]awslib.AwsResource, len(sdkKeyValueStores))
	for _, distribution := range sdkKeyValueStores {
		result = append(result, KeyValueStore{
			KeyValueStore: distribution,
			awsAccount:    accountFromARN(distribution.ARN),
		})
	}

	p.log.Debug("CloudFrontProvider.DescribeKeyValueStores got %d items", len(result))
	return result, nil
}

func (d KeyValueStore) GetRegion() string {
	return ""
}

func (d KeyValueStore) GetResourceArn() string {
	return pointers.Deref(d.ARN)
}

func (d KeyValueStore) GetResourceName() string {
	elements := strings.Split(pointers.Deref(d.ARN), "/")
	return elements[len(elements)-1]
}

func (d KeyValueStore) GetResourceType() string {
	return fetching.CloudFrontKeyValueStoreType
}

func accountFromARN(arn *string) string {
	if arn == nil {
		return ""
	}
	arnString := pointers.Deref(arn)
	if arnString == "" {
		return ""
	}
	elements := strings.Split(arnString, ":")
	if len(elements) != 6 {
		return ""
	}
	return elements[4]
}
