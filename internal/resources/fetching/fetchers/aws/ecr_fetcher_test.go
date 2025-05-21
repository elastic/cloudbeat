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
	"regexp"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/providers/awslib/ecr"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type EcrFetcherTestSuite struct {
	suite.Suite
	resourceCh chan fetching.ResourceInfo
}

type describeRepoMockParameters struct {
	EcrRepositories      []types.Repository
	ExpectedRepositories []string
}

func TestEcrFetcherTestSuite(t *testing.T) {
	s := new(EcrFetcherTestSuite)

	suite.Run(t, s)
}

func (s *EcrFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *EcrFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *EcrFetcherTestSuite) TestCreateFetcher() {
	t := s.T()
	firstRepositoryName := "cloudbeat"
	secondRepositoryName := "cloudbeat1"
	privateRepoWithSlash := "build/cloudbeat"
	repoArn := "arn:aws:ecr:us-west-2:012345678910:repository/ubuntu"

	tests := []struct {
		identityAccount                     string
		namespace                           string
		containers                          []v1.Container
		privateRepositoriesResponseByRegion map[string]describeRepoMockParameters
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
					EcrRepositories: []types.Repository{
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
						},
					},
				},
			},
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
					EcrRepositories: []types.Repository{
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
						},
					},
				},
			},
			[]string{privateRepoWithSlash, secondRepositoryName},
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
					EcrRepositories: []types.Repository{
						{
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &firstRepositoryName,
							RepositoryUri:              nil,
						},
					},
				},
				"us-east-1": {
					ExpectedRepositories: []string{"cloudbeat1"},
					EcrRepositories: []types.Repository{
						{
							RepositoryArn:              &repoArn,
							ImageScanningConfiguration: nil,
							RepositoryName:             &secondRepositoryName,
							RepositoryUri:              nil,
						},
					},
				},
			},
			[]string{firstRepositoryName, secondRepositoryName},
		},
	}
	for _, test := range tests {
		// Need to add services
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
		_, err := kubeclient.CoreV1().Pods(test.namespace).Create(t.Context(), pods, metav1.CreateOptions{})
		s.Require().NoError(err)

		ecrProvider := &ecr.MockRepositoryDescriber{}
		// Init private repositories provider

		ecrProvider.EXPECT().DescribeRepositories(mock.Anything, mock.Anything, mock.Anything).Call.
			Return(func(_ context.Context, repoNames []string, region string) []types.Repository {
				response, ok := test.privateRepositoriesResponseByRegion[region]
				s.True(ok)
				s.Equal(response.ExpectedRepositories, repoNames)

				return response.EcrRepositories
			},
				func(_ context.Context, _ []string, _ string) error {
					return nil
				})

		privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, test.identityAccount)

		privateEcrExecutor := PodDescriber{
			FilterRegex: regexp.MustCompile(privateRepoRegex),
			Provider:    ecrProvider,
		}

		ecrFetcher := EcrFetcher{
			log:          testhelper.NewLogger(s.T()),
			kubeClient:   kubeclient,
			PodDescriber: privateEcrExecutor,
			resourceCh:   s.resourceCh,
		}

		ctx := t.Context()
		err = ecrFetcher.Fetch(ctx, cycle.Metadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Equal(len(test.expectedRepositoriesNames), len(results))
		s.Require().NoError(err)

		// DO NOT REMOVE THIS METHOD
		// The results provided by the channel can return the results in a different order
		// This method order the results before asserting on them
		sort.SliceStable(results, func(i, j int) bool {
			return *results[i].Resource.(EcrResource).RepositoryName <= *results[j].Resource.(EcrResource).RepositoryName
		})

		for i, name := range test.expectedRepositoriesNames {
			ecrResource := results[i].Resource.(EcrResource)
			metadata, err := ecrResource.GetMetadata()
			s.Require().NoError(err)
			s.Equal(name, *ecrResource.RepositoryName)
			s.Equal(*ecrResource.RepositoryName, metadata.Name)
			s.Equal(*ecrResource.RepositoryArn, metadata.ID)
		}
	}
}

func (s *EcrFetcherTestSuite) TestCreateFetcherErrorCases() {
	tests := []struct {
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
		// Need to add services
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
		t := s.T()
		_, err := kubeclient.CoreV1().Pods(test.namespace).Create(t.Context(), pods, metav1.CreateOptions{})
		s.Require().NoError(err)

		ecrProvider := &ecr.MockRepositoryDescriber{}
		ecrProvider.EXPECT().DescribeRepositories(mock.Anything, mock.Anything, mock.Anything).Return([]types.Repository{}, test.error)

		privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, test.identityAccount)

		privateEcrExecutor := PodDescriber{
			FilterRegex: regexp.MustCompile(privateRepoRegex),
			Provider:    ecrProvider,
		}
		ecrFetcher := EcrFetcher{
			log:          testhelper.NewLogger(s.T()),
			kubeClient:   kubeclient,
			PodDescriber: privateEcrExecutor,
			resourceCh:   s.resourceCh,
		}

		ctx := t.Context()
		err = ecrFetcher.Fetch(ctx, cycle.Metadata{})
		s.Require().NoError(err)
		results := testhelper.CollectResources(s.resourceCh)
		s.Empty(results)
	}
}

func (s *EcrFetcherTestSuite) TestEcrResource_GetMetadata() {
	r := EcrResource{
		Repository: ecr.Repository{
			RepositoryArn:  aws.String("test-ecr-arn"),
			RepositoryName: aws.String("test-ecr-name"),
		},
	}
	meta, err := r.GetMetadata()
	s.Require().NoError(err)
	s.Equal(fetching.ResourceMetadata{ID: "test-ecr-arn", Type: "container-registry", SubType: "aws-ecr", Name: "test-ecr-name"}, meta)
	m, err := r.GetElasticCommonData()
	s.Require().NoError(err)
	s.Len(m, 1)
	s.Contains(m, "cloud.service.name")
}
