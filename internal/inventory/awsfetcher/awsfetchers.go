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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/dataprovider/providers/cloud"
	"github.com/elastic/cloudbeat/internal/inventory"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ec2"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/iam"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/s3"
)

func New(logger *logp.Logger, identity *cloud.Identity, cfg aws.Config) []inventory.AssetFetcher {
	iamProvider := iam.NewIAMProvider(logger, cfg, &awslib.MultiRegionClientFactory[iam.AccessAnalyzerClient]{})
	ec2Provider := ec2.NewEC2Provider(logger, identity.Account, cfg, &awslib.MultiRegionClientFactory[ec2.Client]{})
	s3Provider := s3.NewProvider(logger, cfg, &awslib.MultiRegionClientFactory[s3.Client]{}, identity.Account)

	return []inventory.AssetFetcher{
		newEc2InstancesFetcher(logger, identity, ec2Provider),
		newS3BucketFetcher(logger, identity, s3Provider),
		newIamUserFetcher(logger, identity, iamProvider),
		newIamRoleFetcher(logger, identity, iamProvider),
		newIamPolicyFetcher(logger, identity, iamProvider),
	}
}
