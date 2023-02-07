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

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type ElasticCompute interface {
	DescribeNetworkAcl(ctx context.Context) ([]awslib.AwsResource, error)
	DescribeSecurityGroups(ctx context.Context) ([]awslib.AwsResource, error)
	DescribeVPCs(ctx context.Context) ([]awslib.AwsResource, error)
	GetEbsEncryptionByDefault(ctx context.Context) (*EBSEncryption, error)
}

func NewEC2Provider(log *logp.Logger, awsAccountID string, cfg aws.Config) *Provider {
	svc := ec2.NewFromConfig(cfg)
	return &Provider{
		log:          log,
		client:       svc,
		awsAccountID: awsAccountID,
		awsRegion:    cfg.Region,
	}
}
