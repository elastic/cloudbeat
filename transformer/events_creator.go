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
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/resources/fetching"
)

type Transformer struct {
	log           *logp.Logger
	eventMetadata common.MapStr
	commonData    CommonDataInterface
}

func NewTransformer(log *logp.Logger, cd CommonDataInterface, index string) Transformer {
	eventMetadata := common.MapStr{libevents.FieldMetaIndex: index}

	return Transformer{
		log:           log,
		eventMetadata: eventMetadata,
		commonData:    cd,
	}
}

func (t *Transformer) CreateBeatEvents(ctx context.Context, eventData evaluator.EventData) ([]beat.Event, error) {
	if len(eventData.Findings) == 0 {
		return nil, nil
	}

	events := make([]beat.Event, 0)
	resMetadata := eventData.GetMetadata()
	resMetadata.ID = t.commonData.GetResourceId(resMetadata)

	timestamp := time.Now()
	resource := fetching.ResourceFields{
		ResourceMetadata: resMetadata,
		Raw:              eventData.RuleResult.Resource,
	}

	for _, finding := range eventData.Findings {
		event := beat.Event{
			Meta:      t.eventMetadata,
			Timestamp: timestamp,
			Fields: common.MapStr{
				"resource":    resource,
				"resource_id": resMetadata.ID,   // Deprecated - kept for BC
				"type":        resMetadata.Type, // Deprecated - kept for BC
				"cycle_id":    eventData.CycleId,
				"result":      finding.Result,
				"rule":        finding.Rule,
			},
		}

		events = append(events, event)
	}

	return events, nil
}
