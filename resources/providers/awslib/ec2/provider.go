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

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
)

type Provider struct {
	log          *logp.Logger
	client       Client
	awsAccountID string
	awsRegion    string
}

type Client interface {
	DescribeNetworkAcls(ctx context.Context, params *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error)
}

func (p *Provider) DescribeNeworkAcl(ctx context.Context) ([]awslib.AwsResource, error) {
	allAcls := []types.NetworkAcl{}
	input := ec2.DescribeNetworkAclsInput{}
	for {
		output, err := p.client.DescribeNetworkAcls(ctx, &input)
		if err != nil {
			return nil, err
		}
		allAcls = append(allAcls, output.NetworkAcls...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	result := []awslib.AwsResource{}
	for _, nacl := range allAcls {
		result = append(result, NACLInfo{nacl, p.awsAccountID, p.awsRegion})
	}
	return result, nil
}
