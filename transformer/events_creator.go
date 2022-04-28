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
	"fmt"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
)

type Transformer struct {
	context       context.Context
	eval          evaluator.Evaluator
	eventMetadata common.MapStr
	events        []beat.Event
}

func NewTransformer(ctx context.Context, eval evaluator.Evaluator, index string) Transformer {
	eventMetadata := common.MapStr{libevents.FieldMetaIndex: index}
	events := make([]beat.Event, 0)

	return Transformer{
		context:       ctx,
		eval:          eval,
		eventMetadata: eventMetadata,
		events:        events,
	}
}

func (c *Transformer) ProcessAggregatedResources(resources manager.ResourceMap, metadata CycleMetadata) []beat.Event {
	c.events = make([]beat.Event, 0)
	for fetcherType, fetcherResults := range resources {
		c.processEachResource(fetcherResults, ResourceTypeMetadata{CycleMetadata: metadata, Type: fetcherType})
	}

	return c.events
}

func (c *Transformer) processEachResource(results []fetching.Resource, metadata ResourceTypeMetadata) {
	for _, result := range results {
		rid, err := result.GetID()
		if err != nil {
			logp.Error(fmt.Errorf("could not get resource ID, Error: %v", err))
			return
		}
		resMetadata := ResourceMetadata{ResourceTypeMetadata: metadata, ResourceId: rid}
		if err := c.createBeatEvents(result, resMetadata); err != nil {
			logp.Error(fmt.Errorf("failed to create beat events for, %v, Error: %v", metadata, err))
		}
	}
}

func (c *Transformer) createBeatEvents(fetchedResource fetching.Resource, metadata ResourceMetadata) error {
	fetcherResult := fetching.Result{Type: metadata.Type, Resource: fetchedResource.GetData()}
	result, err := c.eval.Decision(c.context, fetcherResult)

	if err != nil {
		logp.Error(fmt.Errorf("error running the policy: %w", err))
		return err
	}

	findings, err := c.eval.Decode(result)
	if err != nil {
		return err
	}

	timestamp := time.Now()
	resource := ResourceFields{
		ID:   metadata.ResourceId,
		Type: metadata.Type,
		Raw:  fetcherResult.Resource,
	}

	for _, finding := range findings {
		event := beat.Event{
			Meta:      c.eventMetadata,
			Timestamp: timestamp,
			Fields: common.MapStr{
				"resource":    resource,
				"resource_id": metadata.ResourceId, // Deprecated - kept for BC
				"type":        metadata.Type,       // Deprecated - kept for BC
				"cycle_id":    metadata.CycleId,
				"result":      finding.Result,
				"rule":        finding.Rule,
			},
		}

		c.events = append(c.events, event)
	}
	return nil
}
