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

package eks

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/providers/awslib"
	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestProvider_DescribeClusters(t *testing.T) {
	t.Run("list clusters error", func(t *testing.T) {
		m := &MockClient{}
		m.On("ListClusters", mock.Anything, mock.Anything).Return(nil, errors.New("boom"))
		p := &Provider{log: testhelper.NewLogger(t), clients: map[string]Client{awslib.DefaultRegion: m}}

		got, err := p.DescribeClusters(t.Context())
		require.Error(t, err)
		assert.Empty(t, got)
	})

	t.Run("lists and describes clusters", func(t *testing.T) {
		m := &MockClient{}
		m.On("ListClusters", mock.Anything, mock.Anything).Return(&eks.ListClustersOutput{
			Clusters: []string{"c1"},
		}, nil)
		m.On("DescribeCluster", mock.Anything, mock.Anything).Return(&eks.DescribeClusterOutput{
			Cluster: &types.Cluster{
				Name:            pointers.Ref("c1"),
				Arn:             pointers.Ref("arn:aws:eks:us-east-1:123:cluster/c1"),
				Status:          types.ClusterStatusActive,
				Version:         pointers.Ref("1.29"),
				PlatformVersion: pointers.Ref("eks.5"),
				Tags:            map[string]string{"Owner": "team-infra"},
				ResourcesVpcConfig: &types.VpcConfigResponse{
					EndpointPublicAccess:  true,
					EndpointPrivateAccess: false,
				},
			},
		}, nil)
		p := &Provider{log: testhelper.NewLogger(t), clients: map[string]Client{awslib.DefaultRegion: m}}

		got, err := p.DescribeClusters(t.Context())
		require.NoError(t, err)
		require.Len(t, got, 1)

		c, ok := got[0].(Cluster)
		require.True(t, ok)
		assert.Equal(t, "c1", c.Name)
		assert.Equal(t, "ACTIVE", c.Status)
		assert.Equal(t, "1.29", c.Version)
		assert.True(t, c.EndpointPublicAccess)
		assert.False(t, c.EndpointPrivateAccess)
		assert.Equal(t, "team-infra", c.GetOwnerTag())
		assert.Equal(t, awslib.DefaultRegion, c.GetRegion())
	})
}
