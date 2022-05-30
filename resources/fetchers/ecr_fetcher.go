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
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

//const PrivateRepoRegexTemplate = "^%s\\.dkr\\.ecr\\.%s\\.amazonaws\\.com\\/([-\\w\\.\\/)[:,@]?"

const PrivateRepoRegexTemplate = "^%s\\.dkr\\.ecr\\.%s\\.amazonaws\\.com\\/([-\\w\\.\\/]+)[:,@]?"
const PublicRepoRegex = "public\\.ecr\\.aws\\/\\w+\\/([-\\w\\.\\/]+)\\:?"

type ECRFetcher struct {
	cfg                      ECRFetcherConfig
	kubeClient               k8s.Interface
	ECRRepositoriesExecutors []ECRExecutor
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
	logp.L().Debug("ecr fetcher starts to fetch data")
	podsAwsImages, err := f.getAwsPodImages(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve pod's aws repositories: %w", err)
	}
	return f.GetRepositoriesEcrDescription(ctx, podsAwsImages), nil
}

func (f *ECRFetcher) GetRepositoriesEcrDescription(ctx context.Context, podsAwsImages []string) []fetching.Resource {
	results := make([]fetching.Resource, 0)
	repositories := make([]ecr.Repository, 0)
	for _, repositoriesExecutor := range f.ECRRepositoriesExecutors {
		ecrRepositories, err := repositoriesExecutor.Execute(ctx, podsAwsImages)
		if err != nil {
			logp.Error(fmt.Errorf("could retrieve ECR repositories: %w", err))
			continue
		}
		repositories = append(repositories, ecrRepositories...)
	}
	results = append(results, ECRResource{repositories})

	return results
}

func (f *ECRFetcher) getAwsPodImages(ctx context.Context) ([]string, error) {
	podsList, err := f.kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("failed to get pods  - %w", err))
		return nil, err
	}

	images := make([]string, 0)
	for _, pod := range podsList.Items {
		for _, container := range pod.Spec.Containers {
			image := container.Image
			// Takes only aws images
			for _, matcher := range f.ECRRepositoriesExecutors {
				if matcher.IsValid(image) {
					images = append(images, image)
					break
				}
			}
		}
	}
	return images, nil
}

func (res ECRResource) GetData() interface{} {
	return res
}

func (res ECRResource) GetMetadata() fetching.ResourceMetadata {
	//TODO implement me
	return fetching.ResourceMetadata{
		ID:      "",
		Type:    "",
		SubType: "",
		Name:    "",
	}
}
