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
	"github.com/elastic/cloudbeat/evaluator"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/manager"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

type Transformer struct {
	context       context.Context
	log           *logp.Logger
	eval          evaluator.Evaluator
	eventMetadata mapstr.M
	commonData    CommonDataInterface
}

func NewTransformer(ctx context.Context, log *logp.Logger, eval evaluator.Evaluator, commonData CommonDataInterface, index string) Transformer {
	eventMetadata := mapstr.M{libevents.FieldMetaIndex: index}

	return Transformer{
		context:       ctx,
		log:           log,
		eval:          eval,
		eventMetadata: eventMetadata,
		commonData:    commonData,
	}
}

func (c *Transformer) ProcessAggregatedResources(resources manager.ResourceMap, cycleMetadata CycleMetadata) []beat.Event {
	events := make([]beat.Event, 0)
	for key, fetcherResults := range resources {
		c.log.Infof("Processing fetched data for resource key %s with %d resources", key, len(fetcherResults))
		arr := c.processEachResource(fetcherResults, cycleMetadata)
		events = append(events, arr...)
	}

	return events
}

func (c *Transformer) processEachResource(results []fetching.Resource, cycleMetadata CycleMetadata) []beat.Event {
	events := make([]beat.Event, 0)
	for _, result := range results {
		arr, err := c.createResourceEvents(result, cycleMetadata)
		if err != nil {
			c.log.Errorf("Failed to create beat events for Cycle ID: %v, error: %v",
				cycleMetadata.CycleId, err)
		} else {
			events = append(events, arr...)
		}
	}

	return events
}

func (c *Transformer) createResourceEvents(fetchedResource fetching.Resource, cycleMetadata CycleMetadata) ([]beat.Event, error) {
	resMetadata := fetchedResource.GetMetadata()
	resMetadata.ID = c.commonData.GetResourceId(resMetadata)
	fetcherResult := fetching.Result{Type: resMetadata.Type, Resource: fetchedResource.GetData()}

	result, err := c.eval.Decision(c.context, fetcherResult)
	if err != nil {
		c.log.Errorf("Error running the policy: %v", err)
		return nil, err
	}

	c.log.Debugf("Eval decision for input: %+v -- %+v", fetcherResult, result)

	findings, err := c.eval.Decode(result)
	if err != nil {
		return nil, err
	}

	c.log.Debugf("Created %d findings for input: %+v", len(findings), fetcherResult)

	timestamp := time.Now()
	resource := fetching.ResourceFields{
		ResourceMetadata: resMetadata,
		Raw:              fetcherResult.Resource,
	}

	events := make([]beat.Event, 0)
	for _, finding := range findings {
		event := beat.Event{
			Meta:      c.eventMetadata,
			Timestamp: timestamp,
			Fields: mapstr.M{
				"resource":    resource,
				"resource_id": resMetadata.ID,   // Deprecated - kept for BC
				"type":        resMetadata.Type, // Deprecated - kept for BC
				"cycle_id":    cycleMetadata.CycleId,
				"result":      finding.Result,
				"rule":        finding.Rule,
			},
		}

		events = append(events, event)
	}

	c.log.Debugf("Created %d events for input: %v", len(events), fetcherResult)
	return events, nil
}
