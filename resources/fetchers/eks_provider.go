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

package fetchers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

type EKSProvider struct {
	client *eks.Client
}

func NewEksProvider(cfg aws.Config) *EKSProvider {
	svc := eks.New(cfg)
	return &EKSProvider{
		client: svc,
	}
}

func (provider EKSProvider) DescribeCluster(ctx context.Context, clusterName string) (*eks.DescribeClusterResponse, error) {
	input := &eks.DescribeClusterInput{
		Name: &clusterName,
	}
	req := provider.client.DescribeClusterRequest(input)
	response, err := req.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster %s from eks , error - %w", clusterName, err)
	}

	return response, err
}
