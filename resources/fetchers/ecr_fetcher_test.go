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
	"github.com/elastic/beats/v7/libbeat/common/kubernetes"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"reflect"
	"regexp"
	"testing"
)

const (
	privateRepositoryTemplate = "^%s\\.dkr\\.ecr\\.%s\\.amazonaws\\.com\\/([-\\w]+)[:,@]?"
	publicRepositoryRegex     = "public\\.ecr\\.aws\\/\\w+\\/([\\w-]+)\\:?"
)

type ECRFetcherTestSuite struct {
	suite.Suite
}

func TestECRFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(ECRFetcherTestSuite))
}

func (s *ECRFetcherTestSuite) SetupTest() {

}

func (s *ECRFetcherTestSuite) TestCreateFetcher() {
	firstRepositoryName := "cloudbeat"
	secondRepositoryName := "cloudbeat1"
	var tests = []struct {
		identityAccount         string
		region                  string
		tag                     string
		namespace               string
		containers              []v1.Container
		expectedRepository      []ecr.Repository
		expectedRepositoryNames []string
	}{
		{
			"704479110758",
			"us-east-2",
			"latest",
			"my-namespace",
			[]v1.Container{
				{
					Image: "704479110758.dkr.ecr.us-east-2.amazonaws.com/cloudbeat:latest",
					Name:  "cloudbeat",
				},
				{
					Image: "704479110758.dkr.ecr.us-east-2.amazonaws.com/cloudbeat1:latest",
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
			[]string{firstRepositoryName, secondRepositoryName},
		},
		{
			"704479110758",
			"us-east-2",
			"latest",
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

		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		// Needs to use the same services
		ecrProvider := &awslib.MockedEcrRepositoryDescriber{}
		ecrProvider.EXPECT().DescribeRepositories(mock.Anything, mock.MatchedBy(func(repo []string) bool {
			return reflect.DeepEqual(test.expectedRepositoryNames, repo)
		})).Return(test.expectedRepository, nil)

		privateRepoRegex := fmt.Sprintf(privateRepositoryTemplate, test.identityAccount, test.region)
		//Maybe will need to change this texts
		regexTexts := []string{
			privateRepoRegex,
			publicRepositoryRegex,
		}
		regexMatchers := []*regexp.Regexp{
			regexp.MustCompile(regexTexts[0]),
			regexp.MustCompile(regexTexts[1]),
		}

		expectedResource := ECRResource{test.expectedRepository}

		ecrFetcher := ECRFetcher{
			cfg:               ECRFetcherConfig{},
			ecrProvider:       ecrProvider,
			kubeClient:        kubeclient,
			repoRegexMatchers: regexMatchers,
		}

		ctx := context.Background()

		result, err := ecrFetcher.Fetch(ctx)
		s.Nil(err)
		s.Equal(1, len(result))

		elbResource := result[0].(ECRResource)
		s.Equal(expectedResource, elbResource)
		s.Equal(len(test.expectedRepositoryNames), len(elbResource.EcrRepositories))

		for i, name := range test.expectedRepositoryNames {
			s.Contains(*elbResource.EcrRepositories[i].RepositoryName, name)
		}
	}

}

func (s *ECRFetcherTestSuite) TestCreateFetcherErrorCases() {
	var tests = []struct {
		identityAccount string
		region          string
		tag             string
		namespace       string
		containers      []v1.Container
		error
	}{
		{
			"704479110758",
			"us-east-2",
			"latest",
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

		mockedKubernetesClientGetter := &providers.MockedKubernetesClientGetter{}
		mockedKubernetesClientGetter.EXPECT().GetClient(mock.Anything, mock.Anything).Return(kubeclient, nil)

		// Needs to use the same services
		ecrProvider := &awslib.MockedEcrRepositoryDescriber{}
		ecrProvider.EXPECT().DescribeRepositories(mock.Anything, mock.Anything).Return(nil, test.error)

		privateRepoRegex := fmt.Sprintf(privateRepositoryTemplate, test.identityAccount, test.region)
		//Maybe will need to change this texts
		regexTexts := []string{
			privateRepoRegex,
			publicRepositoryRegex,
		}
		regexMatchers := []*regexp.Regexp{
			regexp.MustCompile(regexTexts[0]),
			regexp.MustCompile(regexTexts[1]),
		}

		ecrFetcher := ECRFetcher{
			cfg:               ECRFetcherConfig{},
			ecrProvider:       ecrProvider,
			kubeClient:        kubeclient,
			repoRegexMatchers: regexMatchers,
		}

		ctx := context.Background()

		result, err := ecrFetcher.Fetch(ctx)
		s.Nil(result)
		s.NotNil(err)
		s.EqualError(err, fmt.Sprintf("could retrieve ECR repositories: %s", test.error.Error()))

	}
}
