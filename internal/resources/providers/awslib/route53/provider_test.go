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

package route53

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/elastic/cloudbeat/internal/resources/utils/pointers"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
)

func TestProvider_ListRecords(t *testing.T) {
	t.Run("list hosted zones error", func(t *testing.T) {
		m := &MockClient{}
		m.On("ListHostedZones", mock.Anything, mock.Anything).Return(nil, errors.New("boom"))
		p := &Provider{log: testhelper.NewLogger(t), client: m}

		got, err := p.ListRecords(t.Context())
		require.Error(t, err)
		assert.Empty(t, got)
	})

	t.Run("maps records and cleans zone id", func(t *testing.T) {
		m := &MockClient{}
		m.On("ListHostedZones", mock.Anything, mock.Anything).Return(&route53.ListHostedZonesOutput{
			HostedZones: []types.HostedZone{
				{Id: pointers.Ref("/hostedzone/Z123"), Name: pointers.Ref("example.com.")},
			},
		}, nil)
		m.On("ListResourceRecordSets", mock.Anything, mock.Anything).Return(&route53.ListResourceRecordSetsOutput{
			ResourceRecordSets: []types.ResourceRecordSet{
				{
					Name:          pointers.Ref("www.example.com."),
					Type:          types.RRTypeA,
					Weight:        pointers.Ref(int64(10)),
					Region:        types.ResourceRecordSetRegionUsEast1,
					HealthCheckId: pointers.Ref("hc-1"),
					ResourceRecords: []types.ResourceRecord{
						{Value: pointers.Ref("203.0.113.10")},
					},
				},
				{
					Name: pointers.Ref("alias.example.com."),
					Type: types.RRTypeA,
					AliasTarget: &types.AliasTarget{
						DNSName:      pointers.Ref("target.example.com."),
						HostedZoneId: pointers.Ref("Z456"),
					},
				},
			},
		}, nil)
		p := &Provider{log: testhelper.NewLogger(t), client: m}

		got, err := p.ListRecords(t.Context())
		require.NoError(t, err)
		require.Len(t, got, 2)

		first, ok := got[0].(Record)
		require.True(t, ok)
		assert.Equal(t, "Z123", first.ZoneID)
		assert.Equal(t, "example.com.", first.ZoneName)
		assert.Equal(t, "A", first.Type)
		assert.Equal(t, []string{"203.0.113.10"}, first.Records)
		assert.Equal(t, int64(10), *first.Weight)
		assert.Equal(t, "hc-1", first.HealthCheckID)
		assert.Equal(t, "arn:aws:route53:::hostedzone/Z123/recordset/www.example.com./A", first.GetResourceArn())
		assert.Equal(t, "global", first.GetRegion())

		second, ok := got[1].(Record)
		require.True(t, ok)
		assert.Equal(t, "target.example.com.", second.AliasTargetDNS)
		assert.Equal(t, "Z456", second.AliasTargetZoneID)
	})
}
