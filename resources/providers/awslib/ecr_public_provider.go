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
)

// TODO Ofir - https://github.com/elastic/security-team/issues/4035

type ECRPublicProvider struct {
}

func NewEcrPublicProvider() *ECRPublicProvider {
	return &ECRPublicProvider{}
}

// DescribeAllEcrRepositories This method will return a maximum of 100 repository
// If we will ever wish to change it, DescribeRepositories returns results in paginated manner

func (provider *ECRPublicProvider) DescribeAllEcrRepositories(_ context.Context) (EcrRepositories, error) {
	// When repoNames is nil, it will describe all the existing repositories
	var emptyArray = make(EcrRepositories, 0)
	return emptyArray, nil
}

// DescribeRepositories This method will return a maximum of 100 repository
// If we will ever wish to change it, DescribeRepositories returns results in paginated manner
// When repoNames is nil, it will describe all the existing repositories
func (provider *ECRPublicProvider) DescribeRepositories(_ context.Context, _ []string) (EcrRepositories, error) {
	var emptyArray = make(EcrRepositories, 0)
	return emptyArray, nil
}
