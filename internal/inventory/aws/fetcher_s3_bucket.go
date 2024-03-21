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

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
)

type S3BucketFetcher struct {
	logger      *logp.Logger
	provider    s3BucketProvider
	AccountId   string
	AccountName string
}

var s3BucketClassification = inventory.AssetClassification{
	Category:    inventory.CategoryInfrastructure,
	SubCategory: inventory.SubCategoryStorage,
	Type:        inventory.TypeObjectStorage,
	SubType:     inventory.SubTypeS3,
}

type s3BucketProvider interface {
	DescribeBuckets(ctx context.Context) ([]awslib.AwsResource, error)
}

func NewS3BucketFetcher(logger *logp.Logger, identity *cloud.Identity, cfg aws.Config) inventory.AssetFetcher {
	provider := s3.NewProvider(logger, cfg, &awslib.MultiRegionClientFactory[s3.Client]{}, identity.Account)
	return &S3BucketFetcher{
		logger:      logger,
		provider:    provider,
		AccountId:   identity.Account,
		AccountName: identity.AccountAlias,
	}
}

func (s *S3BucketFetcher) Fetch(ctx context.Context, assetChannel chan<- inventory.AssetEvent) {
	s.logger.Info("Fetching S3 Bucket")
	defer s.logger.Info("Fetching S3 Bucket - Finished")

	awsBuckets, err := s.provider.DescribeBuckets(ctx)
	if err != nil {
		s.logger.Errorf("Could not list s3 buckets: %v", err)
		if len(awsBuckets) == 0 {
			return
		}
	}

	buckets := lo.Map(awsBuckets, func(item awslib.AwsResource, _ int) s3.BucketDescription {
		return item.(s3.BucketDescription)
	})

	for _, bucket := range buckets {
		assetChannel <- inventory.NewAssetEvent(
			s3BucketClassification,
			bucket.GetResourceArn(),
			bucket.GetResourceName(),

			inventory.WithRawAsset(bucket),
			inventory.WithCloud(inventory.AssetCloud{
				Provider: inventory.AwsCloudProvider,
				Region:   bucket.Region,
				Account: inventory.AssetCloudAccount{
					Id:   s.AccountId,
					Name: s.AccountName,
				},
				Service: &inventory.AssetCloudService{
					Name: "AWS S3",
				},
			}),
			inventory.WithResourcePolicies(getBucketPolicies(bucket)...),
		)
	}
}

func getBucketPolicies(bucket s3.BucketDescription) []inventory.AssetResourcePolicy {
	if len(bucket.BucketPolicy) == 0 {
		return nil
	}

	version, hasVersion := bucket.BucketPolicy["Version"].(string)
	if !hasVersion {
		version = ""
	}

	switch statements := bucket.BucketPolicy["Statement"].(type) {
	case []map[string]any:
		return convertStatements(statements, version)
	case []any:
		return convertAnyStatements(statements, version)
	case map[string]any:
		return []inventory.AssetResourcePolicy{convertStatement(statements, &version)}
	}
	return nil
}

func convertAnyStatements(statements []any, version string) []inventory.AssetResourcePolicy {
	policies := make([]inventory.AssetResourcePolicy, 0, len(statements))
	for _, statement := range statements {
		policies = append(policies, convertStatement(statement.(map[string]any), &version))
	}
	return policies
}

func convertStatements(statements []map[string]any, version string) []inventory.AssetResourcePolicy {
	policies := make([]inventory.AssetResourcePolicy, 0, len(statements))
	for _, statement := range statements {
		policies = append(policies, convertStatement(statement, &version))
	}
	return policies
}

func convertStatement(statement map[string]any, version *string) inventory.AssetResourcePolicy {
	p := inventory.AssetResourcePolicy{}
	p.Version = version

	if sid, ok := statement["Sid"]; ok {
		p.Id = pointers.Ref(sid.(string))
	}

	if effect, ok := statement["Effect"]; ok {
		p.Effect = effect.(string)
	}

	if anyPrincipal, ok := statement["Principal"]; ok {
		switch principal := anyPrincipal.(type) {
		case string:
			p.Principal = map[string]any{principal: principal}
		case map[string]any:
			p.Principal = principal
		}
	}

	if action, ok := statement["Action"]; ok {
		p.Action = anyToSliceString(action)
	}

	if notAction, ok := statement["NotAction"]; ok {
		p.NotAction = anyToSliceString(notAction)
	}

	if resource, ok := statement["Resource"]; ok {
		p.Resource = anyToSliceString(resource)
	}

	if noResource, ok := statement["NoResource"]; ok {
		p.NoResource = anyToSliceString(noResource)
	}

	if condition, ok := statement["Condition"]; ok {
		p.Condition = condition.(map[string]any)
	}

	return p
}

func anyToSliceString(anyString any) []string {
	switch s := anyString.(type) {
	case string:
		return []string{s}
	case []string:
		return s
	}

	return nil
}
