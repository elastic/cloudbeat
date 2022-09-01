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
	"testing"
	"time"

	"github.com/elastic/beats/v7/libbeat/ecs"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	now        = time.Now()
	resourceID = uuid.Must(uuid.NewV4()).String()
	message    = "message example"

	// fileExample = Fields{
	// 	File:       ecs.File{},
	// 	Event:      buildECSEvent(5, now),
	// 	ResourceID: resourceID,
	// 	Resource:   fetching.ResourceFields{},
	// 	Type:       fetchers.FSResourceType,
	// 	Result:     evaluator.Result{},
	// 	Rule:       evaluator.Rule{},
	// 	Message:    message,
	// }
	// k8sObjectExample = Fields{
	// 	Event:      buildECSEvent(5, now),
	// 	ResourceID: resourceID,
	// 	Resource:   fetching.ResourceFields{},
	// 	Type:       fetchers.K8sObjType,
	// 	Result:     evaluator.Result{},
	// 	Rule:       evaluator.Rule{},
	// 	Message:    message,
	// }
)

func TestFields_MarshalMapStr(t *testing.T) {
	generateUUID = func() string {
		return "feb44eeb-7698-43e8-a308-80e827bb4cd7"
	}
	tests := []struct {
		name     string
		fields   Fields
		expected mapstr.M
		wantErr  bool
	}{
		{
			name: "Type: process",
			fields: Fields{
				Process: ecs.Process{
					PID:         1,
					CommandLine: "cmd",
					Args:        []string{"a", "b"},
					ArgsCount:   2,
					Name:        "proc name",
					Title:       "proc name",
					PGID:        1,
					Parent:      &ecs.Process{PID: 0},
					Start:       now,
					Uptime:      10,
				},
				Event:      buildECSEvent(int64(5), now),
				ResourceID: resourceID,
				Resource: fetching.ResourceFields{
					ResourceMetadata: fetching.ResourceMetadata{
						Type:      fetchers.ProcessResourceType,
						SubType:   fetchers.ProcessSubType,
						Name:      "kubelet",
						ECSFormat: fetchers.ProcessResourceType,
						ID:        resourceID,
					},
					Raw: nil, // TODO: fill
				},
				Type:    fetchers.ProcessType,
				Result:  evaluator.Result{},
				Rule:    evaluator.Rule{},
				Message: message,
			},
			expected: mapstr.M{
				"resource_id": resourceID,
				"event": mapstr.M{
					"category": []string{ecsCategoryConfiguration},
					"created":  now,
					"id":       generateUUID(),
					"kind":     "state",
					"sequence": int64(5),
					"outcome":  ecsOutcomeSuccess,
					"type":     []string{ecsTypeInfo},
				},
				"type":    fetchers.ProcessType,
				"message": "message example",
				"process": mapstr.M{
					"pid":          int64(1),
					"command_line": "cmd",
					"args":         []string{"a", "b"},
					"args_count":   int64(2),
					"name":         "proc name",
					"title":        "proc name",
					"pgid":         int64(1),
					"parent.pid":   int64(0), // TODO: test how it passed in main
					"start":        now,
					"uptime":       int64(10),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := mapstr.M{}
			err := tt.fields.MarshalMapStr(res)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, res)
		})
	}
}
