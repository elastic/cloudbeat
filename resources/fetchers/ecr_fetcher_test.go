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
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
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
	testPrivateRepositoryTemplate = "^%s\\.dkr\\.ecr\\.%s\\.amazonaws\\.com\\/([-\\w]+)[:,@]?"
	testPublicRepositoryRegex     = "public\\.ecr\\.aws\\/\\w+\\/([\\w-]+)\\:?"
)

type ECRFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
	errorCh    chan error
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
	s.resourceCh = make(chan fetching.ResourceInfo)
	s.errorCh = make(chan error, 1)
}

func (s *ECRFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
	close(s.errorCh)
}

func (s *ECRFetcherTestSuite) TestCreateFetcher() {
	firstRepositoryName := "cloudbeat"
	secondRepositoryName := "cloudbeat1"
	var tests = []struct {
		identityAccount         string
		region                  string
		namespace               string
		containers              []v1.Container
		expectedRepository      []ecr.Repository
		expectedRepositoryNames []string
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
			[]string{firstRepositoryName, secondRepositoryName},
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
			return reflect.DeepEqual(test.expectedRepositoryNames, repo)
		})).Return(test.expectedRepository, nil)

		privateRepoRegex := fmt.Sprintf(testPrivateRepositoryTemplate, test.identityAccount, test.region)
		//Maybe will need to change this texts
		regexTexts := []string{
			privateRepoRegex,
			testPublicRepositoryRegex,
		}
		regexMatchers := []*regexp.Regexp{
			regexp.MustCompile(regexTexts[0]),
			regexp.MustCompile(regexTexts[1]),
		}

		ecrFetcher := ECRFetcher{
			log:               s.log,
			cfg:               ECRFetcherConfig{},
			ecrProvider:       ecrProvider,
			kubeClient:        kubeclient,
			repoRegexMatchers: regexMatchers,
			resourceCh:        s.resourceCh,
		}

		ctx := context.Background()
		go func(ch chan error) {
			ch <- ecrFetcher.Fetch(ctx, fetching.CycleMetadata{})
		}(s.errorCh)

		results := testhelper.WaitForResources(s.resourceCh, 1, 2)
		elbResource := results[0].Resource.(ECRResource)

		s.Equal(ECRResource{test.expectedRepository}, elbResource)
		s.Equal(len(test.expectedRepositoryNames), len(elbResource.EcrRepositories))
		s.Nil(<-s.errorCh)

		for i, name := range test.expectedRepositoryNames {
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

		privateRepoRegex := fmt.Sprintf(testPrivateRepositoryTemplate, test.identityAccount, test.region)
		//Maybe will need to change this texts
		regexTexts := []string{
			privateRepoRegex,
			testPublicRepositoryRegex,
		}
		regexMatchers := []*regexp.Regexp{
			regexp.MustCompile(regexTexts[0]),
			regexp.MustCompile(regexTexts[1]),
		}

		ecrFetcher := ECRFetcher{
			log:               s.log,
			cfg:               ECRFetcherConfig{},
			ecrProvider:       ecrProvider,
			kubeClient:        kubeclient,
			repoRegexMatchers: regexMatchers,
			resourceCh:        s.resourceCh,
		}

		ctx := context.Background()
		go func(ch chan error) {
			ch <- ecrFetcher.Fetch(ctx, fetching.CycleMetadata{})
		}(s.errorCh)

		results := testhelper.WaitForResources(s.resourceCh, 1, 2)
		s.Nil(results)
		s.EqualError(<-s.errorCh, fmt.Sprintf("could retrieve ECR repositories: %s", test.error.Error()))
	}
}
