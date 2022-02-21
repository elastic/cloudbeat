package transformer

import (
	"context"
	"encoding/json"
	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/cloudbeat/opa"
	"github.com/elastic/cloudbeat/resources"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

type args struct {
	resource resources.ResourceMap
	metadata CycleMetadata
	cb       CB
}

type testAttr struct {
	name    string
	args    args
	wantErr bool
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
	opaResults   opa.RuleResult
	resourcesMap = map[string][]fetchers.FetchedResource{fetchers.FileSystemType: {fetcherResult}}
	ctx          = context.Background()
	events       = make([]beat.Event, 0)
	cycleId      uuid.UUID
)

func TestMain(m *testing.M) {
	// Collect data before the test
	cycleId, _ = uuid.NewV4()
	parseJsonfile(opaResultsFileName, &opaResults)

	m.Run()
}

func TestTransformer_ProcessAggregatedResources(t *testing.T) {
	tests := []testAttr{
		{
			name: "All events propagated as expected",
			args: args{
				resource: resourcesMap,
				metadata: CycleMetadata{CycleId: cycleId},
				cb:       mockCB(opaResults, nil),
			},
			wantErr: false,
		},
		{
			name: "Events should not be created due to an error",
			args: args{
				resource: resourcesMap,
				metadata: CycleMetadata{CycleId: cycleId},
				cb:       mockCB(opaResults, errors.New("policy err")),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformer := NewTransformer(ctx, tt.args.cb, testIndex)
			generatedEvents := transformer.ProcessAggregatedResources(tt.args.resource, tt.args.metadata)

			if tt.wantErr {
				assert.Equal(t, len(generatedEvents), 0)
			}

			for _, event := range generatedEvents {
				assert.Equal(t, cycleId, event.Fields["cycle_id"], "event cycle_id is not correct")
				assert.NotEmpty(t, event.Timestamp, `event timestamp is missing`)
				assert.NotEmpty(t, event.Fields["result"], "event result is missing")
				assert.NotEmpty(t, event.Fields["rule"], "event rule is missing")
				assert.NotEmpty(t, event.Fields["resource"], "event resource is missing")
				assert.NotEmpty(t, event.Fields["resource_id"], "resource id is missing")
				assert.NotEmpty(t, event.Fields["type"], "resource type is missing")
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

// Mock opa decision func
func mockCB(results interface{}, err error) CB {
	return func(ctx context.Context, input interface{}) (interface{}, error) {
		return results, err
	}
}
