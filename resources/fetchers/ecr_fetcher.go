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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
)

const (
	// PrivateRepoRegexTemplate should identify images with an ecr regex template
	// <account-id>.dkr.ecr.<region>.amazonaws.com/<repository-name>
	PrivateRepoRegexTemplate = "^%s\\.dkr\\.ecr\\.([-\\w]+)\\.amazonaws\\.com\\/([-\\w\\.\\/]+)[:,@]?"
	// PublicRepoRegex should identify images with a public ecr regex template
	// public.ecr.aws/<aws-alias>/<repository>
	PublicRepoRegex          = "public\\.ecr\\.aws\\/\\w+\\/([-\\w\\.\\/]+)\\:?"
	EcrRegionRegexGroup      = 1
	PublicEcrImageRegexIndex = 1
	EcrImageRegexGroup       = 2
)

type EcrFetcher struct {
	log           *logp.Logger
	cfg           EcrFetcherConfig
	kubeClient    k8s.Interface
	PodDescribers []PodDescriber
	resourceCh    chan fetching.ResourceInfo
	awsConfig     aws.Config
}

type PodDescriber struct {
	FilterRegex     *regexp.Regexp
	Provider        awslib.EcrRepositoryDescriber
	ExtractRegion   func(describer PodDescriber, repo string) (string, error)
	ImageRegexIndex int
}

type EcrFetcherConfig struct {
	fetching.AwsBaseFetcherConfig `config:",inline"`
	KubeConfig                    string `config:"Kubeconfig"`
}

type EcrResource struct {
	awslib.EcrRepository
}

func (f *EcrFetcher) Stop() {
}

func (f *EcrFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting EcrFetcher.Fetch")

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
				Resource:      EcrResource{awslib.EcrRepository(repository)},
				CycleMetadata: cMetadata,
			}
		}
	}
	return nil
}

func (f *EcrFetcher) describePodImagesRepositories(ctx context.Context, podsList *v1.PodList, describer PodDescriber) (awslib.EcrRepositories, error) {
	regionToReposMap := getAwsRepositories(podsList, describer)
	f.log.Debugf("sending pods to ecrProviders: %v", regionToReposMap)
	awsRepositories := make(awslib.EcrRepositories, 0)
	for region, repositories := range regionToReposMap {
		// Add configuration
		describedRepo, err := describer.Provider.DescribeRepositories(ctx, f.awsConfig, repositories, region)
		if err != nil {
			f.log.Errorf("could not retrieve pod's aws repositories for region %s: %w", region, err)
		} else {
			awsRepositories = append(awsRepositories, describedRepo...)
		}
	}
	return awsRepositories, nil
}

func getAwsRepositories(podsList *v1.PodList, describer PodDescriber) map[string][]string {
	reposByRegion := make(map[string][]string, 0)

	for _, pod := range podsList.Items {
		for _, container := range pod.Spec.Containers {
			image := container.Image

			// Takes only aws images
			regexMatcher := describer.FilterRegex.FindStringSubmatch(image)
			if regexMatcher != nil {
				repository := regexMatcher[describer.ImageRegexIndex]
				region, err := describer.ExtractRegion(describer, image)
				if err != nil {
					logp.Error(err)
					continue
				}
				reposByRegion[region] = append(reposByRegion[region], repository)
			}
		}
	}
	return reposByRegion
}

func ExtractRegionFromEcrImage(describer PodDescriber, image string) (string, error) {
	regexMatcher := describer.FilterRegex.FindStringSubmatch(image)
	if regexMatcher != nil {
		repository := regexMatcher[EcrRegionRegexGroup]
		return repository, nil
	}

	return "", fmt.Errorf("could not extract region, image does not match the aws ecr template")
}

// ExtractRegionFromPublicEcrImage - Currently the public ecr provider is not functional.
// TODO - This method should be change as part of implementing the public ecr - https://github.com/elastic/security-team/issues/4035
func ExtractRegionFromPublicEcrImage(_ PodDescriber, _ string) (string, error) {
	return "", nil
}

func (res EcrResource) GetData() interface{} {
	return res
}

func (res EcrResource) GetMetadata() (fetching.ResourceMetadata, error) {
	if res.RepositoryArn == nil || res.RepositoryName == nil {
		return fetching.ResourceMetadata{}, errors.New("received nil pointer")
	}

	return fetching.ResourceMetadata{
		ID:      *res.RepositoryArn,
		Type:    fetching.CloudContainerRegistry,
		SubType: fetching.EcrType,
		Name:    *res.RepositoryName,
	}, nil
}

func (res EcrResource) GetElasticCommonData() any { return nil }
