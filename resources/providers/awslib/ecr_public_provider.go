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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// TODO Ofir - https://github.com/elastic/security-team/issues/4035

type ECRPublicProvider struct {
}

func NewEcrPublicProvider() *ECRPublicProvider {
	return &ECRPublicProvider{}
}

// DescribeAllECRRepositories This method will return a maximum of 100 repository
/// If we will ever wish to change it, DescribeRepositories returns results in paginated manner

func (provider *ECRPublicProvider) DescribeAllECRRepositories(_ context.Context, _ aws.Config, _ string) (ECRProviderResponse, error) {
	/// When repoNames is nil, it will describe all the existing Repositories
	var emptyArray = make([]ecr.Repository, 0)
	return ECRProviderResponse{
		Repositories: emptyArray,
	}, nil
}

// DescribeRepositories This method will return a maximum of 100 repository
/// If we will ever wish to change it, DescribeRepositories returns results in paginated manner
/// When repoNames is nil, it will describe all the existing Repositories
func (provider *ECRPublicProvider) DescribeRepositories(_ context.Context, _ aws.Config, _ []string, _ string) (ECRProviderResponse, error) {
	var emptyArray = make([]ecr.Repository, 0)
	return ECRProviderResponse{
		Repositories: emptyArray,
	}, nil
}
