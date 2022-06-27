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
	"io/ioutil"
	"os"
	"testing"

	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/gofrs/uuid"
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

var fetcherResult = fetchers.FileSystemResource{
	Name:    "scheduler.conf",
	Mode:    "700",
	Gid:     20,
	Uid:     501,
	Owner:   "root",
	Group:   "root",
	Path:    "/hostfs/etc/kubernetes/scheduler.conf",
	Inode:   "8901",
	SubType: "file",
}

var (
	opaResults evaluator.RuleResult
	ctx        = context.Background()
	cd         = CommonData{
		clusterId: "test-cluster-id",
		nodeId:    "test-node-id",
	}
)

type EventsCreatorTestSuite struct {
	suite.Suite

	log     *logp.Logger
	cycleId uuid.UUID
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
}

func (s *EventsCreatorTestSuite) TestTransformer_ProcessAggregatedResources() {
	tests := []testAttr{
		{
			name: "All events propagated as expected",
			input: evaluator.EventData{
				RuleResult: opaResults,
				ResourceInfo: fetching.ResourceInfo{
					Resource:      fetcherResult,
					CycleMetadata: fetching.CycleMetadata{CycleId: s.cycleId},
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
					CycleMetadata: fetching.CycleMetadata{CycleId: s.cycleId},
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			transformer := NewTransformer(s.log, cd, testIndex)
			generatedEvents, _ := transformer.CreateBeatEvents(ctx, tt.input)

			for _, event := range generatedEvents {
				resource := event.Fields["resource"].(fetching.ResourceFields)
				s.Equal(s.cycleId, event.Fields["cycle_id"], "event cycle_id is not correct")
				s.NotEmpty(event.Timestamp, `event timestamp is missing`)
				s.NotEmpty(event.Fields["result"], "event result is missing")
				s.NotEmpty(event.Fields["rule"], "event rule is missing")
				s.NotEmpty(resource.Raw, "raw resource is missing")
				s.NotEmpty(resource.SubType, "resource sub type is missing")
				s.NotEmpty(resource.ID, "resource ID is missing")
				s.NotEmpty(resource.Type, "resource  type is missing")
				s.NotEmpty(event.Fields["type"], "resource type is missing") // for BC sake
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

	byteValue, err := ioutil.ReadAll(fetcherDataFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, data)
	if err != nil {
		return err
	}
	return nil
}
