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
	"github.com/elastic/cloudbeat/resources/fetching"
	"regexp"
	"sort"
	"testing"

	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

type EcrFetcherTestSuite struct {
	suite.Suite
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

type describeRepoMockParameters struct {
	EcrRepositories      awslib.EcrRepositories
	ExpectedRepositories []string
}

func TestEcrFetcherTestSuite(t *testing.T) {
	s := new(EcrFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_ecr_fetcher_test_suite")
	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *EcrFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *EcrFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *EcrFetcherTestSuite) TestCreateFetcher() {
	firstRepositoryName := "cloudbeat"
	secondRepositoryName := "cloudbeat1"
	privateRepoWithSlash := "build/cloudbeat"
	repoArn := "arn:aws:ecr:us-west-2:012345678910:repository/ubuntu"
	publicRepoName := "build.security/citools"

	var tests = []struct {
		identityAccount                     string
		namespace                           string
		containers                          []v1.Container
		privateRepositoriesResponseByRegion map[string]describeRepoMockParameters
		publicRepositoriesResponseByRegion  map[string]describeRepoMockParameters
		expectedRepositoriesNames           []string
	}{
		{
			"123456789123",
			"my-namespace",
			[]v1.Container{
				{
					Image: "123456789123.dkr.ecr.us-east-2.amazonaws.com/cloudbeat:latest",
					Name:  "cloudbeat",
				},
				{
					Image: "123456789123.dkr.ecr.us-east-2.amazonaws.com/cloudbeat1:latest",
					Name:  "cloudbeat1",
				},
			},
			map[string]describeRepoMockParameters{
				"us-east-2": {
					ExpectedRepositories: []string{"cloudbeat", "cloudbeat1"},
					EcrRepositories: awslib.EcrRepositories{
						{
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &firstRepositoryName,
							RepositoryUri:              nil,
						},
						{
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &secondRepositoryName,
							RepositoryUri:              nil,
						}},
				},
			},
			map[string]describeRepoMockParameters{},
			[]string{firstRepositoryName, secondRepositoryName},
		},
		{
			"123456789123",
			"my-namespace",
			[]v1.Container{
				{
					Image: "123456789123.dkr.ecr.us-east-2.amazonaws.com/build/cloudbeat:latest",
					Name:  "build/cloudbeat",
				},
				{
					Image: "123456789123.dkr.ecr.us-east-2.amazonaws.com/cloudbeat1:latest",
					Name:  "cloudbeat1",
				},
			},
			map[string]describeRepoMockParameters{
				"us-east-2": {
					ExpectedRepositories: []string{"build/cloudbeat", "cloudbeat1"},
					EcrRepositories: awslib.EcrRepositories{
						{
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &privateRepoWithSlash,
							RepositoryUri:              nil,
						}, {
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &secondRepositoryName,
							RepositoryUri:              nil,
						}},
				},
			}, map[string]describeRepoMockParameters{},
			[]string{privateRepoWithSlash, secondRepositoryName},
		},
		{
			"123456789123",
			"my-namespace",
			[]v1.Container{
				{
					Image: "123456789123.dkr.ecr.us-east-2.amazonaws.com/cloudbeat:latest",
					Name:  "cloudbeat",
				},
				{
					Image: "public.ecr.aws/a7d1s9l0/build.security/citools",
					Name:  "build.security/citools",
				},
			},
			map[string]describeRepoMockParameters{
				"us-east-2": {
					ExpectedRepositories: []string{"cloudbeat"},
					EcrRepositories: awslib.EcrRepositories{
						{
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &firstRepositoryName,
							RepositoryUri:              nil,
						}},
				},
			},
			map[string]describeRepoMockParameters{
				"": {
					ExpectedRepositories: []string{"build.security/citools"},
					EcrRepositories: awslib.EcrRepositories{
						{
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &publicRepoName,
							RepositoryUri:              nil,
						}},
				},
			},
			[]string{publicRepoName, firstRepositoryName},
		},
		{
			"123456789123",
			"my-namespace",
			[]v1.Container{
				{
					Image: "cloudbeat",
					Name:  "cloudbeat",
				},
				{
					Image: "cloudbeat1",
					Name:  "cloudbeat1",
				},
			},
			map[string]describeRepoMockParameters{},
			map[string]describeRepoMockParameters{},
			[]string{},
		},
		{
			"123456789123",
			"my-namespace",
			[]v1.Container{
				{
					Image: "123456789123.dkr.ecr.us-east-2.amazonaws.com/cloudbeat:latest",
					Name:  "cloudbeat",
				},
				{
					Image: "123456789123.dkr.ecr.us-east-1.amazonaws.com/cloudbeat1:latest",
					Name:  "cloudbeat1",
				},
			},
			map[string]describeRepoMockParameters{
				"us-east-2": {
					ExpectedRepositories: []string{"cloudbeat"},
					EcrRepositories: awslib.EcrRepositories{
						{
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &firstRepositoryName,
							RepositoryUri:              nil,
						}},
				},
				"us-east-1": {
					ExpectedRepositories: []string{"cloudbeat1"},
					EcrRepositories: awslib.EcrRepositories{
						{
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &secondRepositoryName,
							RepositoryUri:              nil,
						}},
				},
			},
			map[string]describeRepoMockParameters{},
			[]string{firstRepositoryName, secondRepositoryName},
		},
	}
	for _, test := range tests {
		//Need to add services
		kubeclient := k8sfake.NewSimpleClientset()

		pods := &v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "apps/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testing_pod",
				Namespace: test.namespace,
			},
			Spec: kubernetes.PodSpec{
				NodeName:   "testnode",
				Containers: test.containers,
			},
		}
		_, err := kubeclient.CoreV1().Pods(test.namespace).Create(context.Background(), pods, metav1.CreateOptions{})
		s.NoError(err)

		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		ecrProvider := &awslib.MockEcrRepositoryDescriber{}
		// Init private repositories provider

		ecrProvider.EXPECT().DescribeRepositories(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Call.
			Return(func(ctx context.Context, cfg aws.Config, repoNames []string, region string) awslib.EcrRepositories {
				response, ok := test.privateRepositoriesResponseByRegion[region]
				s.True(ok)
				s.Equal(response.ExpectedRepositories, repoNames)

				return response.EcrRepositories
			},
				func(ctx context.Context, cfg aws.Config, repoNames []string, region string) error {
					return nil
				})

		// Init public repositories provider
		ecrPublicProvider := &awslib.MockEcrRepositoryDescriber{}
		ecrPublicProvider.EXPECT().DescribeRepositories(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Call.
			Return(func(ctx context.Context, cfg aws.Config, repoNames []string, region string) awslib.EcrRepositories {
				response, ok := test.publicRepositoriesResponseByRegion[region]
				s.True(ok)
				s.Equal(response.ExpectedRepositories, repoNames)

				return response.EcrRepositories
			},
				func(ctx context.Context, cfg aws.Config, repoNames []string, region string) error {
					return nil
				})

		privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, test.identityAccount)

		privateEcrExecutor := PodDescriber{
			FilterRegex:     regexp.MustCompile(privateRepoRegex),
			Provider:        ecrProvider,
			ExtractRegion:   ExtractRegionFromEcrImage,
			ImageRegexIndex: EcrImageRegexGroup,
		}

		publicEcrExecutor := PodDescriber{
			FilterRegex:     regexp.MustCompile(PublicRepoRegex),
			Provider:        ecrPublicProvider,
			ExtractRegion:   ExtractRegionFromPublicEcrImage,
			ImageRegexIndex: PublicEcrImageRegexIndex,
		}

		ecrFetcher := EcrFetcher{
			log:           s.log,
			cfg:           EcrFetcherConfig{},
			kubeClient:    kubeclient,
			PodDescribers: []PodDescriber{privateEcrExecutor, publicEcrExecutor},
			resourceCh:    s.resourceCh,
		}

		ctx := context.Background()
		err = ecrFetcher.Fetch(ctx, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Equal(len(test.expectedRepositoriesNames), len(results))
		s.NoError(err)

		// DO NOT REMOVE THIS METHOD
		// The results provided by the channel can return the results in a different order
		// This method order the results before asserting on them
		sort.SliceStable(results, func(i, j int) bool {
			return *results[i].Resource.(EcrResource).RepositoryName <= *results[j].Resource.(EcrResource).RepositoryName
		})

		for i, name := range test.expectedRepositoriesNames {
			ecrResource := results[i].Resource.(EcrResource)
			metadata, err := ecrResource.GetMetadata()
			s.NoError(err)
			s.Equal(name, *ecrResource.RepositoryName)
			s.Equal(*ecrResource.RepositoryName, metadata.Name)
			s.Equal(*ecrResource.RepositoryArn, metadata.ID)
		}
	}

}

func (s *EcrFetcherTestSuite) TestCreateFetcherErrorCases() {
	var tests = []struct {
		identityAccount string
		region          string
		namespace       string
		containers      []v1.Container
		error
	}{
		{
			"704479110758",
			"us-east-2",
			"my-namespace",
			[]v1.Container{
				{
					Image: "cloudbeat:latest",
					Name:  "cloudbeat",
				},
			},
			fmt.Errorf("ecr error"),
		},
	}
	for _, test := range tests {
		//Need to add services
		kubeclient := k8sfake.NewSimpleClientset()

		pods := &v1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "apps/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testing_pod",
				Namespace: test.namespace,
			},
			Spec: kubernetes.PodSpec{
				NodeName:   "testnode",
				Containers: test.containers,
			},
		}
		_, err := kubeclient.CoreV1().Pods(test.namespace).Create(context.Background(), pods, metav1.CreateOptions{})
		s.NoError(err)

		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		ecrProvider := &awslib.MockEcrRepositoryDescriber{}
		ecrProvider.EXPECT().DescribeRepositories(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(awslib.EcrRepositories{}, test.error)

		ecrPublicProvider := &awslib.MockEcrRepositoryDescriber{}
		ecrPublicProvider.EXPECT().DescribeRepositories(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(awslib.EcrRepositories{}, test.error)

		privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, test.identityAccount)

		privateEcrExecutor := PodDescriber{
			FilterRegex:     regexp.MustCompile(privateRepoRegex),
			Provider:        ecrProvider,
			ExtractRegion:   ExtractRegionFromEcrImage,
			ImageRegexIndex: EcrImageRegexGroup,
		}

		publicEcrExecutor := PodDescriber{
			FilterRegex:     regexp.MustCompile(PublicRepoRegex),
			Provider:        ecrPublicProvider,
			ExtractRegion:   ExtractRegionFromPublicEcrImage,
			ImageRegexIndex: PublicEcrImageRegexIndex,
		}

		ecrFetcher := EcrFetcher{
			log:           s.log,
			cfg:           EcrFetcherConfig{},
			kubeClient:    kubeclient,
			PodDescribers: []PodDescriber{privateEcrExecutor, publicEcrExecutor},
			resourceCh:    s.resourceCh,
		}

		ctx := context.Background()
		err = ecrFetcher.Fetch(ctx, fetching.CycleMetadata{})
		s.NoError(err)
		results := testhelper.CollectResources(s.resourceCh)
		s.Equal(0, len(results))
	}
}
