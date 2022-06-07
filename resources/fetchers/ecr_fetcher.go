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
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/gofrs/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const PrivateRepoRegexTemplate = "^%s\\.dkr\\.ecr\\.%s\\.amazonaws\\.com\\/([-\\w\\.\\/]+)[:,@]?"
const PublicRepoRegex = "public\\.ecr\\.aws\\/\\w+\\/([-\\w\\.\\/]+)\\:?"

type ECRFetcher struct {
	log           *logp.Logger
	cfg           ECRFetcherConfig
	kubeClient    k8s.Interface
	PodDescribers []PodDescriber
}

type PodDescriber struct {
	DescriberRegex *regexp.Regexp
	Provider       awslib.EcrRepositoryDescriber
}

type ECRFetcherConfig struct {
	fetching.BaseFetcherConfig
	Kubeconfig string `config:"Kubeconfig"`
}

type EcrRepositories []ecr.Repository

type ECRResource struct {
	EcrRepositories
}

func (f *ECRFetcher) Stop() {
}

func (f *ECRFetcher) Fetch(ctx context.Context) ([]fetching.Resource, error) {
	f.log.Debug("Starting ECRFetcher.Fetch")
	results := make([]fetching.Resource, 0)
	repositories := make([]ecr.Repository, 0)
	for _, podDescriber := range f.PodDescribers {
		repo, err := f.getAwsRepositories(ctx, podDescriber)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve pod's aws repositories: %w", err)
		}
		repositories = append(repositories, repo...)
	}
	results = append(results, ECRResource{repositories})

	return results, nil
}

func (f *ECRFetcher) getAwsRepositories(ctx context.Context, describer PodDescriber) ([]ecr.Repository, error) {
	podsList, err := f.kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("failed to get pods  - %w", err))
		return nil, err
	}
	repositories := make([]string, 0)

	for _, pod := range podsList.Items {
		for _, container := range pod.Spec.Containers {
			image := container.Image
			// Takes only aws images
			regexMatcher := describer.DescriberRegex.FindStringSubmatch(image)
			{
				if regexMatcher != nil {
					repository := regexMatcher[1]
					repositories = append(repositories, repository)
				}
			}
		}
	}
	return describer.Provider.DescribeRepositories(ctx, repositories)
}

func (res ECRResource) GetData() interface{} {
	return res
}

func (res ECRResource) GetMetadata() fetching.ResourceMetadata {
	uid, _ := uuid.NewV4()
	return fetching.ResourceMetadata{
		ID:      uid.String(),
		Type:    ECRType,
		SubType: ECRType,
		Name:    "AWS repositories",
	}
}
