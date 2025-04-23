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

package transformer

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/evaluator"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	gcpfetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/gcp"
	fetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/k8s"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/version"
)

type testAttr struct {
	name  string
	input evaluator.EventData
	bdpp  func() dataprovider.CommonDataProvider
	cdpp  func() dataprovider.ElasticCommonDataProvider
	idpp  func() dataprovider.IdProvider
}

const (
	opaResultsFileName = "opa_results.json"
	resourceId         = "test_resource_id"
	enrichedKey        = "enriched_key"
	enrichedValue      = "enrichedValue"
)

var versionInfo = version.CloudbeatVersionInfo{
	Version: version.Version{Version: "test_version"},
}

var fetcherResult = fetchers.FSResource{
	EvalResource: fetchers.EvalFSResource{
		Name:    "scheduler.conf",
		Mode:    "700",
		Gid:     "20",
		Uid:     "501",
		Owner:   "root",
		Group:   "root",
		Path:    "/hostfs/etc/kubernetes/scheduler.conf",
		Inode:   "8901",
		SubType: "file",
	},
	ElasticCommon: fetchers.FileCommonData{
		Name: "scheduler.conf",
	},
}

var (
	opaResults evaluator.RuleResult
	ctx        = context.Background()
)

type EventsCreatorTestSuite struct {
	suite.Suite
}

func TestSuite(t *testing.T) {
	s := new(EventsCreatorTestSuite)
	suite.Run(t, s)
}

func (s *EventsCreatorTestSuite) SetupSuite() {
	fetcherDataFile, err := os.Open(opaResultsFileName)
	s.Require().NoError(err)
	defer func() {
		s.Require().NoError(fetcherDataFile.Close())
	}()

	byteValue, err := io.ReadAll(fetcherDataFile)
	s.Require().NoError(err)

	err = json.Unmarshal(byteValue, &opaResults)
	s.Require().NoError(err)
}

func (s *EventsCreatorTestSuite) TestTransformer_ProcessAggregatedResources() {
	tests := []testAttr{
		{
			name: "All events propagated as expected",
			input: evaluator.EventData{
				RuleResult: opaResults,
				ResourceInfo: fetching.ResourceInfo{
					Resource:      fetcherResult,
					CycleMetadata: cycle.Metadata{},
				},
			},
			bdpp: func() dataprovider.CommonDataProvider {
				dataProviderMock := dataprovider.NewMockCommonDataProvider(s.T())
				mockEnrichEvent := func(event *beat.Event, _ fetching.ResourceMetadata) {
					_, err := event.Fields.Put(enrichedKey, enrichedValue)
					s.Require().NoError(err)
				}
				dataProviderMock.EXPECT().EnrichEvent(mock.Anything, mock.Anything).Run(mockEnrichEvent).Return(nil)
				return dataProviderMock
			},
			cdpp: func() dataprovider.ElasticCommonDataProvider {
				dataProviderMock := dataprovider.NewMockElasticCommonDataProvider(s.T())
				ret := map[string]any{
					"cloudbeat": versionInfo,
				}
				dataProviderMock.EXPECT().GetElasticCommonData().Return(ret, nil)
				return dataProviderMock
			},
			idpp: func() dataprovider.IdProvider {
				idProviderMock := dataprovider.NewMockIdProvider(s.T())
				idProviderMock.EXPECT().GetId(mock.Anything, mock.Anything).Return(resourceId)
				return idProviderMock
			},
		},
		{
			name: "Events should not be created due zero findings",
			input: evaluator.EventData{
				RuleResult: evaluator.RuleResult{
					Findings: []evaluator.Finding{},
					Metadata: evaluator.Metadata{},
					Resource: nil,
				},
				ResourceInfo: fetching.ResourceInfo{
					Resource:      fetcherResult,
					CycleMetadata: cycle.Metadata{},
				},
			},
			bdpp: func() dataprovider.CommonDataProvider {
				dataProviderMock := dataprovider.NewMockCommonDataProvider(s.T())
				return dataProviderMock
			},
			cdpp: func() dataprovider.ElasticCommonDataProvider {
				dataProviderMock := dataprovider.NewMockElasticCommonDataProvider(s.T())
				return dataProviderMock
			},
			idpp: func() dataprovider.IdProvider {
				idProviderMock := dataprovider.NewMockIdProvider(s.T())
				return idProviderMock
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cdp := tt.cdpp()
			bdp := tt.bdpp()
			idp := tt.idpp()

			transformer := NewTransformer(testhelper.NewLogger(s.T()), &config.Config{}, bdp, cdp, idp)
			generatedEvents, _ := transformer.CreateBeatEvents(ctx, tt.input)

			for _, event := range generatedEvents {
				resource := event.Fields["resource"].(fetching.ResourceFields)
				s.NotEmpty(event.Timestamp, `event timestamp is missing`)
				s.NotEmpty(event.Fields["result"], "event result is missing")
				s.NotEmpty(event.Fields["rule"], "event rule is missing")
				s.NotEmpty(event.Fields["file"], "elastic common data is missing")
				s.NotEmpty(event.Fields["related"], "related data is missing")
				s.NotEmpty(resource.Raw, "raw resource is missing")
				s.NotEmpty(resource.SubType, "resource sub type is missing")
				s.Equal("test_resource_id", resource.ID)
				s.NotEmpty(resource.Type, "resource  type is missing")
				s.NotEmpty(event.Fields["event"], "resource event is missing")
				s.Equal(event.Fields["cloudbeat"], versionInfo)
				s.Equal(enrichedValue, event.Fields[enrichedKey])
				s.Regexp("^Rule \".*\": (passed|failed)$", event.Fields["message"], "event message is not correct")
			}
		})
	}
}

func (s *EventsCreatorTestSuite) TestTransformer_GetPreferredRawValue() {
	tests := []struct {
		name     string
		expected any
		input    fetching.Resource
	}{
		{
			name:     "returns nil for GcpLoggingAsset",
			input:    &gcpfetchers.GcpLoggingAsset{},
			expected: nil,
		},
		{
			name:     "returns nil for GcpMonitoringAsset",
			input:    &gcpfetchers.GcpMonitoringAsset{},
			expected: nil,
		},
		{
			name:     "returns nil for GcpPoliciesAsset",
			input:    &gcpfetchers.GcpPoliciesAsset{},
			expected: nil,
		},
		{
			name:     "returns nil for GcpServiceUsageAsset",
			input:    &gcpfetchers.GcpServiceUsageAsset{},
			expected: nil,
		},
		{
			name:     "returns the actual resource when the resource is not a gcp-wrapper-resource",
			input:    &gcpfetchers.GcpAsset{},
			expected: &gcpfetchers.GcpAsset{},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			eventData := evaluator.EventData{
				RuleResult: evaluator.RuleResult{
					Findings: []evaluator.Finding{},
					Metadata: evaluator.Metadata{},
					Resource: test.input,
				},
				ResourceInfo: fetching.ResourceInfo{
					Resource:      test.input,
					CycleMetadata: cycle.Metadata{},
				},
			}
			result := getPreferredRawValue(eventData)
			s.Equal(test.expected, result)
		})
	}
}
