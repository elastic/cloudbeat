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
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type EksFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

func TestEksFetcherTestSuite(t *testing.T) {
	s := new(EksFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_eks_fetcher_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *EksFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *EksFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *EksFetcherTestSuite) TestEksFetcherFetch() {
	var tests = []struct {
		clusterName     string
		clusterResponse eks.DescribeClusterResponse
	}{
		{
			"cluster_name",
			eks.DescribeClusterResponse{},
		},
	}

	for _, test := range tests {
		awsConfig := fetching.AwsBaseFetcherConfig{
			BaseFetcherConfig: fetching.BaseFetcherConfig{},
			AwsConfig:         aws.Config{},
		}

		eksConfig := EKSFetcherConfig{
			AwsBaseFetcherConfig: awsConfig,
			ClusterName:          test.clusterName,
		}
		eksProvider := &awslib.MockedEksClusterDescriber{}
		expectedResource := EKSResource{&test.clusterResponse}

		eksProvider.EXPECT().DescribeCluster(mock.Anything, test.clusterName).Return(&test.clusterResponse, nil)
		eksFetcher := EKSFetcher{
			log:         s.log,
			cfg:         eksConfig,
			eksProvider: eksProvider,
			resourceCh:  s.resourceCh,
		}

		ctx := context.Background()
		err := eksFetcher.Fetch(ctx, fetching.CycleMetadata{})

		results := testhelper.CollectResources(s.resourceCh)
		eksResource := results[0].Resource.(EKSResource)

		s.Equal(expectedResource, eksResource)
		s.Nil(err)

	}
}

func (s *EksFetcherTestSuite) TestEksFetcherFetchWhenErrorOccurs() {
	clusterName := "my-cluster"
	awsConfig := fetching.AwsBaseFetcherConfig{
		BaseFetcherConfig: fetching.BaseFetcherConfig{},
		AwsConfig:         aws.Config{},
	}
	eksConfig := EKSFetcherConfig{
		AwsBaseFetcherConfig: awsConfig,
		ClusterName:          clusterName,
	}
	eksProvider := &awslib.MockedEksClusterDescriber{}

	expectedErr := fmt.Errorf("my error")
	eksProvider.EXPECT().DescribeCluster(mock.Anything, clusterName).Return(nil, expectedErr)
	eksFetcher := EKSFetcher{
		log:         s.log,
		cfg:         eksConfig,
		eksProvider: eksProvider,
		resourceCh:  s.resourceCh,
	}

	ctx := context.Background()
	err := eksFetcher.Fetch(ctx, fetching.CycleMetadata{})
	results := testhelper.CollectResources(s.resourceCh)
	s.Equal(0, len(results))
	s.Equal(expectedErr, err)
}
