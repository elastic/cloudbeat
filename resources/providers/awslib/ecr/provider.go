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

package ecr

import (
	"context"
	"fmt"
	ecrClient "github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
)

// DescribeAllEcrRepositories returns a list of all the existing repositories
func (provider *Provider) DescribeAllEcrRepositories(ctx context.Context, region string) ([]types.Repository, error) {
	/// When repoNames is nil, it will describe all the existing Repositories
	return provider.DescribeRepositories(ctx, nil, region)
}

// DescribeRepositories returns a list of repositories that match the given names.
// When repoNames is empty, it will describe all the existing repositories
func (provider *Provider) DescribeRepositories(ctx context.Context, repoNames []string, region string) ([]types.Repository, error) {
	if len(repoNames) == 0 {
		return nil, nil
	}

	client, err := awslib.GetClient(&region, provider.clients)
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	input := &ecrClient.DescribeRepositoriesInput{
		RepositoryNames: repoNames,
	}

	var repos []types.Repository
	for {
		output, err := client.DescribeRepositories(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("error describing repositories: %w", err)
		}

		repos = append(repos, output.Repositories...)

		// Check if there are more repositories to retrieve
		if output.NextToken == nil {
			break
		}

		input.NextToken = output.NextToken
	}

	return repos, nil
}
