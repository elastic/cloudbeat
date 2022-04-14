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
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type EksFetcherTestSuite struct {
	suite.Suite
	factory fetching.Factory
}

func TestEksFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(EksFetcherTestSuite))
}

func (s *EksFetcherTestSuite) SetupTest() {

}

func (s *EksFactoryTestSuite) TestEksFetcherFetch() {
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
		eksConfig := EKSFetcherConfig{
			BaseFetcherConfig: fetching.BaseFetcherConfig{},
			ClusterName:       test.clusterName,
		}
		eksProvider := &awslib.MockedEksClusterDescriber{}
		expectedResource := EKSResource{&test.clusterResponse}

		eksProvider.EXPECT().DescribeCluster(mock.Anything, test.clusterName).Return(&test.clusterResponse, nil)
		eksFetcher := EKSFetcher{
			cfg:         eksConfig,
			eksProvider: eksProvider,
		}

		ctx := context.Background()
		result, err := eksFetcher.Fetch(ctx)
		s.Nil(err)

		eksResource := result[0].(EKSResource)
		s.Equal(expectedResource, eksResource)
	}
}

func (s *EksFactoryTestSuite) TestEksFetcherFetchWhenErrorOccurs() {
	clusterName := "my-cluster"
	eksConfig := EKSFetcherConfig{
		BaseFetcherConfig: fetching.BaseFetcherConfig{},
		ClusterName:       clusterName,
	}
	eksProvider := &awslib.MockedEksClusterDescriber{}

	expectedErr := fmt.Errorf("my error")
	eksProvider.EXPECT().DescribeCluster(mock.Anything, clusterName).Return(nil, expectedErr)
	eksFetcher := EKSFetcher{
		cfg:         eksConfig,
		eksProvider: eksProvider,
	}

	ctx := context.Background()
	_, err := eksFetcher.Fetch(ctx)
	s.NotNil(err)
	s.Equal(expectedErr, err)
}
