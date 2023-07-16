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

package inventory

import (
	context "context"
	"testing"

	"cloud.google.com/go/asset/apiv1/assetpb"
	iampb "cloud.google.com/go/iam/apiv1/iampb"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/suite"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	gcplib "github.com/elastic/cloudbeat/resources/providers/gcplib/auth"
)

type ProviderTestSuite struct {
	suite.Suite
}

func TestInventoryProviderTestSuite(t *testing.T) {
	s := new(ProviderTestSuite)

	suite.Run(t, s)
}

func (s *ProviderTestSuite) TestListAllAssetTypesByName() {
	ctx := context.Background()
	mockIterator := new(MockIterator)
	gcpClientWrapper := &GcpClientWrapper{
		Close: func() error { return nil },
		ListAssets: func(ctx context.Context, req *assetpb.ListAssetsRequest, opts ...gax.CallOption) Iterator {
			return mockIterator
		},
	}
	provider := &Provider{
		log:    logp.NewLogger("test"),
		client: gcpClientWrapper,
		ctx:    ctx,
		Config: gcplib.GcpFactoryConfig{
			ProjectId:  "1",
			ClientOpts: []option.ClientOption{},
		},
	}

	mockIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", Resource: &assetpb.Resource{}}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName2", Resource: &assetpb.Resource{}}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{Name: "AssetName1", IamPolicy: &iampb.Policy{}}, nil).Once()
	mockIterator.On("Next").Return(&assetpb.Asset{}, iterator.Done).Once()

	value, err := provider.ListAllAssetTypesByName([]string{"test"})
	assetNames := Map(value, func(asset *assetpb.Asset) string { return asset.Name })
	resourceAssets := Filter(value, func(asset *assetpb.Asset) bool { return asset.Resource != nil })
	policyAssets := Filter(value, func(asset *assetpb.Asset) bool { return asset.IamPolicy != nil })

	s.Assert().NoError(err)
	s.Assert().Equal(ContainsString(assetNames, "AssetName1"), true)
	s.Assert().Equal(len(resourceAssets), 2) // 2 assets with resources (assetName1, assetName2)
	s.Assert().Equal(len(policyAssets), 1)   // 1 assets with policy 	(assetName1)
	s.Assert().Equal(len(value), 2)          // 2 assets in total 		(assetName1 merged resource/policy, assetName2)
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func Filter[T any](ts []T, fn func(T) bool) []T {
	var result []T
	for _, t := range ts {
		if fn(t) {
			result = append(result, t)
		}
	}
	return result
}

func ContainsString(arr []string, value string) bool {
	return Filter(arr, func(s string) bool { return s == value }) != nil
}
