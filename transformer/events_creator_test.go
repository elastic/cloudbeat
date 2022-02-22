package transformer

import (
	"context"
	"encoding/json"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/resources"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

type args struct {
	resource resources.ResourceMap
	metadata CycleMetadata
}

type testAttr struct {
	name    string
	args    args
	wantErr bool
	mocks   []MethodMock
}

type MethodMock struct {
	methodName string
	args       []interface{}
	returnArgs []interface{}
}

const (
	opaResultsFileName = "opa_results.json"
	testIndex          = "test_index"
)

var fetcherResult = fetchers.FileSystemResource{
	FileName: "scheduler.conf",
	FileMode: "700",
	Gid:      "root",
	Uid:      "root",
	Path:     "/hostfs/etc/kubernetes/scheduler.conf",
	Inode:    "8901",
}

var (
	opaResults      evaluator.RuleResult
	mockedEvaluator = evaluator.MockedEvaluator{}
	resourcesMap    = map[string][]fetchers.FetchedResource{fetchers.FileSystemType: {fetcherResult}}
	ctx             = context.Background()
)

type EventsCreatorTestSuite struct {
	suite.Suite
	cycleId         uuid.UUID
	mockedEvaluator evaluator.MockedEvaluator
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(EventsCreatorTestSuite))
}

func (s *EventsCreatorTestSuite) SetupSuite() {
	parseJsonfile(opaResultsFileName, &opaResults)
	s.cycleId, _ = uuid.NewV4()
}

func (s *EventsCreatorTestSuite) SetupTest() {
	s.mockedEvaluator = evaluator.MockedEvaluator{}
}

func (s *EventsCreatorTestSuite) TestTransformer_ProcessAggregatedResources() {
	var tests = []testAttr{
		{
			name: "All events propagated as expected",
			args: args{
				resource: resourcesMap,
				metadata: CycleMetadata{CycleId: s.cycleId},
			},
			mocks: []MethodMock{{
				methodName: "Decision",
				args:       []interface{}{ctx, mock.AnythingOfType("FetcherResult")},
				returnArgs: []interface{}{mock.Anything, nil},
			}, {
				methodName: "Decode",
				args:       []interface{}{ctx, mock.Anything},
				returnArgs: []interface{}{opaResults.Findings, nil},
			},
			},
			wantErr: false,
		},
		{
			name: "Events should not be created due to a policy error",
			args: args{
				resource: resourcesMap,
				metadata: CycleMetadata{CycleId: s.cycleId},
			},
			mocks: []MethodMock{{
				methodName: "Decision",
				args:       []interface{}{ctx, mock.AnythingOfType("FetcherResult")},
				returnArgs: []interface{}{mock.Anything, errors.New("policy err")},
			}, {
				methodName: "Decode",
				args:       []interface{}{ctx, mock.Anything},
				returnArgs: []interface{}{opaResults.Findings, nil},
			},
			},
			wantErr: true,
		},
		{
			name: "Events should not be created due to a parse error",
			args: args{
				resource: resourcesMap,
				metadata: CycleMetadata{CycleId: s.cycleId},
			},
			mocks: []MethodMock{{
				methodName: "Decision",
				args:       []interface{}{ctx, mock.AnythingOfType("FetcherResult")},
				returnArgs: []interface{}{mock.Anything, nil},
			}, {
				methodName: "Decode",
				args:       []interface{}{ctx, mock.Anything},
				returnArgs: []interface{}{nil, errors.New("parse err")},
			},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		s.SetupTest()
		s.Run(tt.name, func() {
			for _, methodMock := range tt.mocks {
				s.mockedEvaluator.On(methodMock.methodName, methodMock.args...).Return(methodMock.returnArgs...)
			}

			transformer := NewTransformer(ctx, s.mockedEvaluator, testIndex)
			generatedEvents := transformer.ProcessAggregatedResources(tt.args.resource, tt.args.metadata)

			if tt.wantErr {
				s.Equal(0, len(generatedEvents))
			}

			for _, event := range generatedEvents {
				s.Equal(s.cycleId, event.Fields["cycle_id"], "event cycle_id is not correct")
				s.NotEmpty(event.Timestamp, `event timestamp is missing`)
				s.NotEmpty(event.Fields["result"], "event result is missing")
				s.NotEmpty(event.Fields["rule"], "event rule is missing")
				s.NotEmpty(event.Fields["resource"], "event resource is missing")
				s.NotEmpty(event.Fields["resource_id"], "resource id is missing")
				s.NotEmpty(event.Fields["type"], "resource type is missing")
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

	json.Unmarshal(byteValue, data)
	return nil
}
