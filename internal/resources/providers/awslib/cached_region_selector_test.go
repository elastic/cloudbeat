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
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

type CachedRegionSelectorTestSuite struct {
	suite.Suite
}

func TestCachedRegionSelectorTestSuite(t *testing.T) {
	testhelper.SkipLong(t)

	s := new(CachedRegionSelectorTestSuite)

	suite.Run(t, s)
}

func (s *CachedRegionSelectorTestSuite) initTest() (*cachedRegionSelector, *MockRegionsSelector) {
	var err error
	mocked := &MockRegionsSelector{}
	selector := newCachedRegionSelector(mocked, s.T().Name(), time.Second) // Unique cache space for each test
	s.Require().NoError(err)
	return selector, mocked
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_SingleCall() {
	selector, mocked := s.initTest()
	mocked.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil)
	result, err := selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_DoubleCallCached() {
	selector, mocked := s.initTest()
	mocked.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil)
	result, err := selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)

	time.Sleep(10 * time.Millisecond)

	result, err = selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)

	mocked.AssertNumberOfCalls(s.T(), "Regions", 1)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_DoubleCallEvicted() {
	selector, mocked := s.initTest()
	mocked.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil)
	result, err := selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)

	time.Sleep(1 * time.Second)

	result, err = selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
	mocked.AssertNumberOfCalls(s.T(), "Regions", 2)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_CacheEvictionFlow() {
	selector, mocked := s.initTest()
	mocked.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil)

	result, err := selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
	mocked.AssertNumberOfCalls(s.T(), "Regions", 1)

	time.Sleep(1 * time.Second)
	result, err = selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
	mocked.AssertNumberOfCalls(s.T(), "Regions", 2)

	time.Sleep(20 * time.Millisecond)
	result, err = selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
	mocked.AssertNumberOfCalls(s.T(), "Regions", 2)

	time.Sleep(1 * time.Second)
	result, err = selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
	mocked.AssertNumberOfCalls(s.T(), "Regions", 3)

	time.Sleep(20 * time.Millisecond)
	result, err = selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
	mocked.AssertNumberOfCalls(s.T(), "Regions", 3)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_FirstFail() {
	selector, mocked := s.initTest()
	mocked.EXPECT().Regions(mock.Anything, mock.Anything).Return(nil, errors.New("mock")).Once()
	result, err := selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().Error(err)
	s.Empty(result)

	mocked.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil).Once()
	result, err = selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{usRegion, euRegion}, result)
	mocked.AssertNumberOfCalls(s.T(), "Regions", 2)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_ParallelCalls() {
	selector, mocked := s.initTest()
	mocked.EXPECT().Regions(mock.Anything, mock.Anything).Return(successfulOutput, nil)
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			time.Sleep(time.Duration(i*5) * time.Millisecond)
			result, err := selector.Regions(context.Background(), *awssdk.NewConfig())
			s.Require().NoError(err)
			s.Equal([]string{usRegion, euRegion}, result)
		}(i)
	}

	wg.Wait()
	mocked.AssertNumberOfCalls(s.T(), "Regions", 1)
}

func (s *CachedRegionSelectorTestSuite) TestCachedRegionSelector_ParallelCallsFail() {
	selector, mocked := s.initTest()
	mocked.EXPECT().Regions(mock.Anything, mock.Anything).Return(nil, errors.New("mock"))
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			time.Sleep(time.Duration(i*5) * time.Millisecond)
			result, err := selector.Regions(context.Background(), *awssdk.NewConfig())
			s.Require().Error(err)
			s.Empty(result)
		}(i)
	}

	wg.Wait()
	mocked.AssertNumberOfCalls(s.T(), "Regions", 5)
}
