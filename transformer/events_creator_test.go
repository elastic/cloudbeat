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
	"github.com/elastic/cloudbeat/dataprovider"
	"github.com/elastic/cloudbeat/version"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/stretchr/testify/suite"
)

type testAttr struct {
	name  string
	input evaluator.EventData
}

const (
	opaResultsFileName = "opa_results.json"
	testIndex          = "test_index"
)

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
	log              *logp.Logger
	dataProviderMock dataprovider.MockCommonDataInterface
}

func TestSuite(t *testing.T) {
	s := new(EventsCreatorTestSuite)
	s.log = logp.NewLogger("cloudbeat_events_creator_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(t, s)
}

func (s *EventsCreatorTestSuite) SetupSuite() {
	err := parseJsonfile(opaResultsFileName, &opaResults)
	if err != nil {
		s.log.Errorf("Could not parse JSON file: %v", err)
		return
	}

	s.dataProviderMock = dataprovider.MockCommonDataInterface{}
}

func (s *EventsCreatorTestSuite) TestTransformer_ProcessAggregatedResources() {
	tests := []testAttr{
		{
			name: "All events propagated as expected",
			input: evaluator.EventData{
				RuleResult: opaResults,
				ResourceInfo: fetching.ResourceInfo{
					Resource:      fetcherResult,
					CycleMetadata: fetching.CycleMetadata{},
				},
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
					CycleMetadata: fetching.CycleMetadata{},
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			clusterName := "test_cluster_name"
			s.dataProviderMock.EXPECT().GetResourceId(mock.Anything).Return("test_resource_id")
			s.dataProviderMock.EXPECT().GetVersionInfo().Return(version.CloudbeatVersionInfo{
				Version: version.Version{Version: "test_version"},
			})
			s.dataProviderMock.EXPECT().GetClusterName().Return(clusterName)

			transformer := NewTransformer(s.log, &s.dataProviderMock, testIndex)
			generatedEvents, _ := transformer.CreateBeatEvents(ctx, tt.input)

			for _, event := range generatedEvents {
				eventClusterName, _ := event.GetValue("orchestrator.cluster.name")
				s.Equal(clusterName, eventClusterName)
				resource := event.Fields["resource"].(fetching.ResourceFields)
				s.NotEmpty(event.Timestamp, `event timestamp is missing`)
				s.NotEmpty(event.Fields["result"], "event result is missing")
				s.NotEmpty(event.Fields["rule"], "event rule is missing")
				s.NotEmpty(event.Fields["file"], "elastic common data is missing")
				s.NotEmpty(resource.Raw, "raw resource is missing")
				s.NotEmpty(resource.SubType, "resource sub type is missing")
				s.NotEmpty(resource.ID, "resource ID is missing")
				s.NotEmpty(resource.Type, "resource  type is missing")
				s.NotEmpty(event.Fields["event"], "resource event is missing")
				s.NotEmpty(event.Fields["cloudbeat"], "cloudbeat's version info is missing")
				s.Regexp(regexp.MustCompile("^Rule \".*\": (passed|failed)$"), event.Fields["message"], "event message is not correct")
			}
		})
	}
}
func (s *EventsCreatorTestSuite) TestTransformer_EmptyClusterName() {
	tests := []testAttr{
		{
			name: "All events propagated as expected",
			input: evaluator.EventData{
				RuleResult: opaResults,
				ResourceInfo: fetching.ResourceInfo{
					Resource:      fetcherResult,
					CycleMetadata: fetching.CycleMetadata{},
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.dataProviderMock.EXPECT().GetResourceId(mock.Anything).Return("test_resource_id")
			s.dataProviderMock.EXPECT().GetVersionInfo().Return(version.CloudbeatVersionInfo{
				Version: version.Version{Version: "test_version"},
			})
			s.dataProviderMock.EXPECT().GetClusterName().Return("")

			transformer := NewTransformer(s.log, &s.dataProviderMock, testIndex)
			generatedEvents, _ := transformer.CreateBeatEvents(ctx, tt.input)

			for _, event := range generatedEvents {
				s.False(event.Fields.HasKey("orchestrator.cluster.name"))
			}
		})
	}
}

func parseJsonfile(filename string, data interface{}) error {
	fetcherDataFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fetcherDataFile.Close()

	byteValue, err := io.ReadAll(fetcherDataFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, data)
	if err != nil {
		return err
	}
	return nil
}
