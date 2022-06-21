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
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
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
	log *logp.Logger
}

func TestECRFetcherTestSuite(t *testing.T) {
	s := new(ECRFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_ecr_fetcher_test_suite")
	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *ECRFetcherTestSuite) TestFetcherFetch() {
	firstRepositoryName := "cloudbeat"
	secondRepositoryName := "cloudbeat1"
	publicRepoName := "build.security/citools"
	privateRepoWithSlash := "build/cloudbeat"

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
				ImageScanningConfiguration: nil,
				RepositoryName:             &firstRepositoryName,
				RepositoryUri:              nil,
			}, {
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
				ImageScanningConfiguration: nil,
				RepositoryName:             &privateRepoWithSlash,
				RepositoryUri:              nil,
			}, {
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
				ImageScanningConfiguration: nil,
				RepositoryName:             &firstRepositoryName,
				RepositoryUri:              nil,
			}}, []ecr.Repository{{
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
		s.Nil(err)

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
		expectedResource := ECRResource{expectedRepositories}

		ecrFetcher := ECRFetcher{
			log:           s.log,
			cfg:           ECRFetcherConfig{},
			kubeClient:    kubeclient,
			PodDescribers: []PodDescriber{privateEcrExecutor, publicEcrExecutor},
		}

		ctx := context.Background()

		result, err := ecrFetcher.Fetch(ctx)
		s.Nil(err)
		s.Equal(1, len(result))

		elbResource := result[0].(ECRResource)
		s.Equal(expectedResource, elbResource)
		s.Equal(len(expectedRepositories), len(elbResource.EcrRepositories))

		for i, name := range test.expectedRepositoriesNames {
			s.Contains(*elbResource.EcrRepositories[i].RepositoryName, name)
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
		s.Nil(err)

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
		}

		ctx := context.Background()

		result, err := ecrFetcher.Fetch(ctx)
		s.Equal(0, len(result))
		s.EqualError(err, "could not retrieve pod's aws repositories: ecr error")
	}
}
