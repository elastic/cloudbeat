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
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
)

type EcrPublicRepository types.Repository
type EcrPublicRepositories []types.Repository

type EcrPublicProvider struct {
}

func NewEcrPublicProvider() *EcrPublicProvider {
	return &EcrPublicProvider{}
}

// DescribeAllEcrRepositories returns a list of all the existing public ecr repositories within this region
func (provider *EcrPublicProvider) DescribeAllEcrRepositories(ctx context.Context, cfg aws.Config, region string) (EcrPublicRepositories, error) {
	// When repoNames is nil, it will describe all the existing repositories
	return provider.DescribeRepositories(ctx, cfg, nil, region)
}

// DescribeRepositories returns a list of public ecr repositories that match the given names.
// When repoNames is empty, it will describe all the existing repositories
func (provider *EcrPublicProvider) DescribeRepositories(ctx context.Context, cfg aws.Config, repoNames []string, region string) (EcrPublicRepositories, error) {
	if len(repoNames) == 0 {
		return EcrPublicRepositories{}, nil
	}

	cfg.Region = region
	svc := ecrpublic.NewFromConfig(cfg)
	input := &ecrpublic.DescribeRepositoriesInput{
		RepositoryNames: repoNames,
	}
	var repos EcrPublicRepositories
	paginator := ecrpublic.NewDescribeRepositoriesPaginator(svc, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error DescribeRepositories with Paginator: %w", err)
		}
		repos = append(repos, page.Repositories...)
	}

	return repos, nil

}
