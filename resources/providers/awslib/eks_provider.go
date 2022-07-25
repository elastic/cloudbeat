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

package awslib

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

type EksClusterOutput eks.DescribeClusterOutput

type EksClusterDescriber interface {
	DescribeCluster(ctx context.Context, clusterName string) (EksClusterOutput, error)
}

type EKSProvider struct {
	client *eks.Client
}

func NewEksProvider(cfg aws.Config) *EKSProvider {
	svc := eks.NewFromConfig(cfg)
	return &EKSProvider{
		client: svc,
	}
}

func (provider EKSProvider) DescribeCluster(ctx context.Context, clusterName string) (EksClusterOutput, error) {
	input := &eks.DescribeClusterInput{
		Name: &clusterName,
	}

	response, err := provider.client.DescribeCluster(ctx, input)
	if err != nil {
		return EksClusterOutput{}, fmt.Errorf("failed to describe cluster %s from eks, error - %w", clusterName, err)
	}

	return EksClusterOutput(*response), err
}
