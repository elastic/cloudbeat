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
	"github.com/elastic/cloudbeat/resources/fetching"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
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

type ECRFetcherTestSuite struct {
	suite.Suite
	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

func TestECRFetcherTestSuite(t *testing.T) {
	s := new(ECRFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_ecr_fetcher_test_suite")
	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ECRFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *ECRFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *ECRFetcherTestSuite) TestCreateFetcher() {
	firstRepositoryName := "cloudbeat"
	secondRepositoryName := "cloudbeat1"
	publicRepoName := "build.security/citools"
	privateRepoWithSlash := "build/cloudbeat"
	repoArn := "arn:aws:ecr:us-west-2:012345678910:repository/ubuntu"

	var tests = []struct {
		identityAccount                 string
		region                          string
		namespace                       string
		containers                      []v1.Container
		expectedRepositories            []ecr.Repository
		expectedPublicRepositories      []ecr.Repository
		expectedRepositoriesNames       []string
		expectedPublicRepositoriesNames []string
	}{
		{
			"123456789123",
			"us-east-2",
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
			[]ecr.Repository{{
				RepositoryArn:              &repoArn,
				ImageScanningConfiguration: nil,
				RepositoryName:             &firstRepositoryName,
				RepositoryUri:              nil,
			}, {
				RepositoryArn:              &repoArn,
				ImageScanningConfiguration: nil,
				RepositoryName:             &secondRepositoryName,
				RepositoryUri:              nil,
			}},
			[]ecr.Repository{},
			[]string{firstRepositoryName, secondRepositoryName},
			[]string{},
		}, {
			"123456789123",
			"us-east-2",
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
			[]ecr.Repository{{
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
			[]ecr.Repository{},
			[]string{privateRepoWithSlash, secondRepositoryName},
			[]string{},
		},
		{
			"123456789123",
			"us-east-2",
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
			[]ecr.Repository{{
				RepositoryArn:              &repoArn,
				ImageScanningConfiguration: nil,
				RepositoryName:             &firstRepositoryName,
				RepositoryUri:              nil,
			}}, []ecr.Repository{{
				RepositoryArn:              &repoArn,
				ImageScanningConfiguration: nil,
				RepositoryName:             &publicRepoName,
				RepositoryUri:              nil,
			}},
			[]string{firstRepositoryName},
			[]string{publicRepoName},
		},
		{
			"123456789123",
			"us-east-2",
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
			[]ecr.Repository{},
			[]ecr.Repository{},
			[]string{},
			[]string{},
		},
		{
			"123456789123",
			"us-east-1",
			"my-namespace",
			[]v1.Container{
				{
					Image: "123456789123.dkr.ecr.wrong-region.amazonaws.com/cloudbeat:latest",
					Name:  "cloudbeat",
				},
				{
					Image: "123456789123.dkr.ecr.wrong-region.amazonaws.com/cloudbeat1:latest",
					Name:  "cloudbeat1",
				},
			},
			[]ecr.Repository{},
			[]ecr.Repository{},
			[]string{},
			[]string{},
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

		// Needs to use the same services
		ecrProvider := &awslib.MockedEcrRepositoryDescriber{}
		ecrProvider.EXPECT().DescribeRepositories(mock.Anything, mock.MatchedBy(func(repo []string) bool {
			return s.Equal(test.expectedRepositoriesNames, repo)
		})).Return(test.expectedRepositories, nil)

		ecrPublicProvider := &awslib.MockedEcrRepositoryDescriber{}
		ecrPublicProvider.EXPECT().DescribeRepositories(mock.Anything, mock.MatchedBy(func(repo []string) bool {
			return s.Equal(test.expectedPublicRepositoriesNames, repo)
		})).Return(test.expectedPublicRepositories, nil)

		privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, test.identityAccount, test.region)

		privateEcrExecutor := PodDescriber{
			FilterRegex: regexp.MustCompile(privateRepoRegex),
			Provider:    ecrProvider,
		}

		publicEcrExecutor := PodDescriber{
			FilterRegex: regexp.MustCompile(PublicRepoRegex),
			Provider:    ecrPublicProvider,
		}

		expectedRepositories := append(test.expectedRepositories, test.expectedPublicRepositories...)
		ecrFetcher := ECRFetcher{
			log:           s.log,
			cfg:           ECRFetcherConfig{},
			kubeClient:    kubeclient,
			PodDescribers: []PodDescriber{privateEcrExecutor, publicEcrExecutor},
			resourceCh:    s.resourceCh,
		}

		ctx := context.Background()
		err = ecrFetcher.Fetch(ctx, fetching.CycleMetadata{})
		results := testhelper.CollectResources(s.resourceCh)

		s.Equal(len(expectedRepositories), len(results))
		s.NoError(err)

		for i, name := range test.expectedRepositoriesNames {
			ecrResource := results[i].Resource.(ECRResource)
			metadata, err := ecrResource.GetMetadata()
			s.NoError(err)
			s.Equal(name, *ecrResource.RepositoryName)
			s.Equal(*ecrResource.RepositoryName, metadata.Name)
			s.Equal(*ecrResource.RepositoryArn, metadata.ID)
		}
	}

}

func (s *ECRFetcherTestSuite) TestCreateFetcherErrorCases() {
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

		// Needs to use the same services
		ecrProvider := &awslib.MockedEcrRepositoryDescriber{}
		ecrProvider.EXPECT().DescribeRepositories(mock.Anything, mock.Anything).Return(nil, test.error)

		ecrPublicProvider := &awslib.MockedEcrRepositoryDescriber{}
		ecrPublicProvider.EXPECT().DescribeRepositories(mock.Anything, mock.Anything).Return(nil, test.error)

		privateRepoRegex := fmt.Sprintf(PrivateRepoRegexTemplate, test.identityAccount, test.region)

		privateEcrExecutor := PodDescriber{
			FilterRegex: regexp.MustCompile(privateRepoRegex),
			Provider:    ecrProvider,
		}

		publicEcrExecutor := PodDescriber{
			FilterRegex: regexp.MustCompile(PublicRepoRegex),
			Provider:    ecrPublicProvider,
		}

		ecrFetcher := ECRFetcher{
			log:           s.log,
			cfg:           ECRFetcherConfig{},
			kubeClient:    kubeclient,
			PodDescribers: []PodDescriber{privateEcrExecutor, publicEcrExecutor},
			resourceCh:    s.resourceCh,
		}

		ctx := context.Background()
		err = ecrFetcher.Fetch(ctx, fetching.CycleMetadata{})

		results := testhelper.CollectResources(s.resourceCh)
		s.Equal(0, len(results))
		s.EqualError(err, "could not retrieve pod's aws repositories: ecr error")
	}
}
