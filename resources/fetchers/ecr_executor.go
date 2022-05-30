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
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"regexp"
)

type ECRExecutor struct {
	regexValidator *regexp.Regexp
	handler        awslib.EcrRepositoryDescriber
}

func (e *ECRExecutor) IsValid(imageName string) bool {
	return e.regexValidator.MatchString(imageName)
}

func (e ECRExecutor) Execute(ctx context.Context, images []string) ([]ecr.Repository, error) {
	repositories := make([]string, 0)
	// Takes only aws images
	for _, image := range images {
		if e.regexValidator.MatchString(image) {
			// Extract the repository name out of the image name
			repository := e.regexValidator.FindStringSubmatch(image)[1]
			repositories = append(repositories, repository)
		}
	}
	return e.handler.DescribeRepositories(ctx, repositories)
}
