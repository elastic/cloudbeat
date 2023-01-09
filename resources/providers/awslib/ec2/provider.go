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
	clients      []Client
	awsAccountID string
	awsRegion    string
}

type Client interface {
	DescribeNetworkAcls(ctx context.Context, params *ec2.DescribeNetworkAclsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkAclsOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
}

func (p *Provider) DescribeNetworkAcl(ctx context.Context) ([]awslib.AwsResource, error) {
	var allAcls []types.NetworkAcl
	for _, client := range p.clients {
		input := ec2.DescribeNetworkAclsInput{}

		for {
			output, err := client.DescribeNetworkAcls(ctx, &input)
			if err != nil {
				return nil, err
			}
			allAcls = append(allAcls, output.NetworkAcls...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}
	}

	var result []awslib.AwsResource
	for _, nacl := range allAcls {
		result = append(result, NACLInfo{nacl, p.awsAccountID, p.awsRegion})
	}
	return result, nil
}

func (p *Provider) DescribeSecurityGroups(ctx context.Context) ([]awslib.AwsResource, error) {
	var all []types.SecurityGroup
	for _, client := range p.clients {
		input := &ec2.DescribeSecurityGroupsInput{}

		for {
			output, err := client.DescribeSecurityGroups(ctx, input)
			if err != nil {
				return nil, err
			}
			all = append(all, output.SecurityGroups...)
			if output.NextToken == nil {
				break
			}
			input.NextToken = output.NextToken
		}
	}

	var result []awslib.AwsResource
	for _, sg := range all {
		result = append(result, SecurityGroup{sg, p.awsAccountID, p.awsRegion})
	}
	return result, nil
}
