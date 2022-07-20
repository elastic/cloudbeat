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
	"github.com/pkg/errors"
	"regexp"

	"github.com/elastic/cloudbeat/resources/providers/awslib"
	v1 "k8s.io/api/core/v1"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	PrivateRepoRegexTemplate = "^%s\\.dkr\\.ecr\\.%s\\.amazonaws\\.com\\/([-\\w\\.\\/]+)[:,@]?"
	PublicRepoRegex          = "public\\.ecr\\.aws\\/\\w+\\/([-\\w\\.\\/]+)\\:?"
	ecrResourceType          = "aws-image-registry"
	ecrResourceSubType       = "ecr"
)

type ECRFetcher struct {
	log           *logp.Logger
	cfg           ECRFetcherConfig
	kubeClient    k8s.Interface
	PodDescribers []PodDescriber
	resourceCh    chan fetching.ResourceInfo
}

type PodDescriber struct {
	FilterRegex *regexp.Regexp
	Provider    awslib.EcrRepositoryDescriber
}

type ECRFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
	KubeConfig                    string `config:"Kubeconfig"`
}

type EcrRepository ecr.Repository

type ECRResource struct {
	EcrRepository
}

func (f *ECRFetcher) Stop() {
}

func (f *ECRFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting ECRFetcher.Fetch")

	podsList, err := f.kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		logp.Error(fmt.Errorf("failed to get pods  - %w", err))
		return err
	}

	for _, podDescriber := range f.PodDescribers {
		ecrDescribedRepositories, err := f.describePodImagesRepositories(ctx, podsList, podDescriber)
		if err != nil {
			return fmt.Errorf("could not retrieve pod's aws repositories: %w", err)
		}
		for _, repository := range ecrDescribedRepositories {
			f.resourceCh <- fetching.ResourceInfo{
				Resource:      ECRResource{EcrRepository(repository)},
				CycleMetadata: cMetadata,
			}
		}
	}

	return nil
}

func (f *ECRFetcher) describePodImagesRepositories(ctx context.Context, podsList *v1.PodList, describer PodDescriber) ([]ecr.Repository, error) {

	repositories := make([]string, 0)

	for _, pod := range podsList.Items {
		for _, container := range pod.Spec.Containers {
			image := container.Image

			// Takes only aws images
			regexMatcher := describer.FilterRegex.FindStringSubmatch(image)
			if regexMatcher != nil {
				repository := regexMatcher[1]
				repositories = append(repositories, repository)
			}
		}
	}

	f.log.Debugf("sending pods to ecrProviders: %v", repositories)

	return describer.Provider.DescribeRepositories(ctx, repositories)
}

func (res ECRResource) GetData() interface{} {
	return res
}

func (res ECRResource) GetMetadata() (fetching.ResourceMetadata, error) {
	if res.RepositoryArn == nil || res.RepositoryName == nil {
		return fetching.ResourceMetadata{}, errors.New("received nil pointer")
	}

	return fetching.ResourceMetadata{
		ID:      *res.RepositoryArn,
		Type:    fetching.CloudContainerRegistry,
		SubType: fetching.ECRType,
		Name:    *res.RepositoryName,
	}, nil
}

func (res ECRResource) GetElasticCommonData() any { return nil }
