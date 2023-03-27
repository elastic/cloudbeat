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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CachedRegionSelectorTestSuite struct {
	suite.Suite
	selector *cachedRegionSelector
	mock     *MockRegionsSelector
}

func TestCachedRegionSelectorTestSuite(t *testing.T) {
	s := new(CachedRegionSelectorTestSuite)

	suite.Run(t, s)
}

func (s *CachedRegionSelectorTestSuite) SetupTest() {
	s.mock = &MockRegionsSelector{}
	s.selector = newCachedRegionSelector(s.mock, &cachedRegions{})
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_SingleCall() {
	s.mock.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil)
	result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_DoubleCallCached() {
	s.mock.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil)
	result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)

	result, err = s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)

	s.mock.AssertNumberOfCalls(s.T(), "Regions", 1)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_FirstFail() {
	s.mock.EXPECT().Regions(mock.Anything, mock.Anything).Return(nil, errors.New("mock")).Once()
	result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Error(err)
	s.Len(result, 0)

	s.mock.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil).Once()
	result, err = s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
	s.mock.AssertNumberOfCalls(s.T(), "Regions", 2)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_ParallelCalls() {
	s.mock.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil)
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
	s.mock.AssertNumberOfCalls(s.T(), "Regions", 1)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_ParallelCallsFail() {

	s.mock.EXPECT().Regions(mock.Anything, mock.Anything).Return(nil, errors.New("mock"))
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
	s.mock.AssertNumberOfCalls(s.T(), "Regions", 5)
}
