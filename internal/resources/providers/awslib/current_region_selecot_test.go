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
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	ec2imds "github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var successfulCurrentCloudRegionOutput = &ec2imds.InstanceIdentityDocument{
	Region: euRegion,
}

type CurrentRegionSelectorTestSuite struct {
	suite.Suite
	selector *currentRegionSelector
	mock     *MockMetadataProvider
}

func TestCurrentRegionSelectorTestSuite(t *testing.T) {
	s := new(CurrentRegionSelectorTestSuite)

	suite.Run(t, s)
}

func (s *CurrentRegionSelectorTestSuite) SetupTest() {
	s.selector = &currentRegionSelector{}
	s.mock = &MockMetadataProvider{}
	s.selector.client = s.mock
}

func (s *CurrentRegionSelectorTestSuite) TestCurrentRegionSelector_SingleCall() {
	s.mock.EXPECT().GetMetadata(mock.Anything, mock.Anything).Return(successfulCurrentCloudRegionOutput, nil)
	result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{euRegion}, result)
}

func (s *CurrentRegionSelectorTestSuite) TestCurrentRegionSelector_FirstFail() {
	s.mock.EXPECT().GetMetadata(mock.Anything, mock.Anything).Return(nil, errors.New("mock")).Once()
	result, err := s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().Error(err)
	s.Empty(result)

	s.mock.EXPECT().GetMetadata(mock.Anything, mock.Anything).Return(successfulCurrentCloudRegionOutput, nil).Once()
	result, err = s.selector.Regions(context.Background(), *awssdk.NewConfig())
	s.Require().NoError(err)
	s.Equal([]string{euRegion}, result)
	s.mock.AssertNumberOfCalls(s.T(), "GetMetadata", 2)
}
