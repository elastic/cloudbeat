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

package awsfetcher

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/elb"
	elbv2 "github.com/elastic/cloudbeat/internal/resources/providers/awslib/elb_v2"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/lambda"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/rds"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/sns"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

func New(ctx context.Context, logger *clog.Logger, identity *cloud.Identity, cfg aws.Config) []inventory.AssetFetcher {
	ec2Provider := ec2.NewEC2Provider(ctx, logger, identity.Account, cfg, &awslib.MultiRegionClientFactory[ec2.Client]{})
	elbProvider := elb.NewElbProvider(ctx, logger, identity.Account, cfg, &awslib.MultiRegionClientFactory[elb.Client]{})
	elbv2Provider := elbv2.NewElbV2Provider(ctx, logger, cfg, &awslib.MultiRegionClientFactory[elbv2.Client]{})
	iamProvider := iam.NewIAMProvider(ctx, logger, cfg, &awslib.MultiRegionClientFactory[iam.AccessAnalyzerClient]{})
	lambdaProvider := lambda.NewLambdaProvider(ctx, logger, cfg, &awslib.MultiRegionClientFactory[lambda.Client]{})
	rdsProvider := rds.NewProvider(ctx, logger, cfg, &awslib.MultiRegionClientFactory[rds.Client]{}, ec2Provider)
	s3Provider := s3.NewProvider(ctx, logger, cfg, &awslib.MultiRegionClientFactory[s3.Client]{}, identity.Account)
	snsProvider := sns.NewSNSProvider(ctx, logger, cfg, &awslib.MultiRegionClientFactory[sns.Client]{})

	return []inventory.AssetFetcher{
		newEc2InstancesFetcher(logger, identity, ec2Provider),
		newElbFetcher(logger, identity, elbProvider, elbv2Provider),
		newIamPolicyFetcher(logger, identity, iamProvider),
		newIamRoleFetcher(logger, identity, iamProvider),
		newIamUserFetcher(logger, identity, iamProvider),
		newLambdaFetcher(logger, identity, lambdaProvider),
		newNetworkingFetcher(logger, identity, ec2Provider),
		newRDSFetcher(logger, identity, rdsProvider),
		newS3BucketFetcher(logger, identity, s3Provider),
		newSNSFetcher(logger, identity, snsProvider),
	}
}
