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

package awslib

import (
	"context"
	"errors"
	"sync"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var successfulDescribeCloudRegionOutput = &ec2.DescribeRegionsOutput{
	Regions: []types.Region{
		{
			RegionName: awssdk.String(usRegion),
		},
		{
			RegionName: awssdk.String(euRegion),
		},
	}}

type AllRegionSelectorTestSuite struct {
	suite.Suite
	selector *allRegionsSelector
	mock     *mockDescribeCloudRegions
}

func TestAllRegionSelectorTestSuite(t *testing.T) {
	s := new(AllRegionSelectorTestSuite)

	suite.Run(t, s)
}

func (s *AllRegionSelectorTestSuite) SetupTest() {
	s.selector = newAllRegionSelector()
	s.mock = &mockDescribeCloudRegions{}
	s.selector.client = s.mock
}

func (s *AllRegionSelectorTestSuite) TestAllRegionSelector_SingleCall() {
	s.mock.EXPECT().DescribeRegions(mock.Anything, mock.Anything).Return(successfulDescribeCloudRegionOutput, nil)
	result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
}

func (s *AllRegionSelectorTestSuite) TestAllRegionSelector_DoubleCallCached() {
	s.mock.EXPECT().DescribeRegions(mock.Anything, mock.Anything).Return(successfulDescribeCloudRegionOutput, nil)
	result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)

	result, err = s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)

	s.mock.AssertNumberOfCalls(s.T(), "DescribeRegions", 1)
}

func (s *AllRegionSelectorTestSuite) TestAllRegionSelector_FirstFail() {
	s.mock.EXPECT().DescribeRegions(mock.Anything, mock.Anything).Return(nil, errors.New("mock")).Once()
	result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Error(err)
	s.Len(result, 0)

	s.mock.EXPECT().DescribeRegions(mock.Anything, mock.Anything).Return(successfulDescribeCloudRegionOutput, nil).Once()
	result, err = s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
	s.mock.AssertNumberOfCalls(s.T(), "DescribeRegions", 2)
}

func (s *AllRegionSelectorTestSuite) TestAllRegionSelector_ParallelCalls() {
	s.mock.EXPECT().DescribeRegions(mock.Anything, mock.Anything).Return(successfulDescribeCloudRegionOutput, nil)
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
			s.NoError(err)
			s.Equal([]string{usRegion, euRegion}, result)
		}()
	}

	wg.Wait()
	s.mock.AssertNumberOfCalls(s.T(), "DescribeRegions", 1)
}

func (s *AllRegionSelectorTestSuite) TestAllRegionSelector_ParallelCallsFail() {

	s.mock.EXPECT().DescribeRegions(mock.Anything, mock.Anything).Return(nil, errors.New("mock"))
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
			s.Error(err)
			s.Len(result, 0)
		}()
	}

	wg.Wait()
	s.mock.AssertNumberOfCalls(s.T(), "DescribeRegions", 5)
}
