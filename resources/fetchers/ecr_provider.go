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
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

type ECRProvider struct {
	client *ecr.Client
}

func NewEcrProvider(cfg aws.Config) *ECRProvider {
	svc := ecr.New(cfg)
	return &ECRProvider{
		client: svc,
	}
}

// DescribeAllECRRepositories This method will return a maximum of 100 repository
/// If we will ever wish to change it, DescribeRepositories returns results in paginated manner
func (provider *ECRProvider) DescribeAllECRRepositories(ctx context.Context) ([]ecr.Repository, error) {
	/// When repoNames is nil, it will describe all the existing repositories
	return provider.DescribeRepositories(ctx, nil)
}

// DescribeRepositories This method will return a maximum of 100 repository
/// If we will ever wish to change it, DescribeRepositories returns results in paginated manner
/// When repoNames is nil, it will describe all the existing repositories
func (provider *ECRProvider) DescribeRepositories(ctx context.Context, repoNames []string) ([]ecr.Repository, error) {
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: repoNames,
	}
	req := provider.client.DescribeRepositoriesRequest(input)
	response, err := req.Send(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository:%s from ecr, error - %w", repoNames, err)
	}

	return response.Repositories, err
}
