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
	context       context.Context
	log           *logp.Logger
	eval          evaluator.Evaluator
	eventMetadata common.MapStr
	events        Events
	commonData    CommonDataInterface
}

func NewTransformer(ctx context.Context, log *logp.Logger, eval evaluator.Evaluator, events Events, commonData CommonDataInterface, index string) Transformer {
	eventMetadata := common.MapStr{libevents.FieldMetaIndex: index}

	return Transformer{
		context:       ctx,
		log:           log,
		eval:          eval,
		events:        events,
		eventMetadata: eventMetadata,
		commonData:    commonData,
	}
}

func (c *Transformer) ProcessAggregatedResources(ctx context.Context, resourceChan <-chan fetching.ResourcesInfo) {
	for {
		select {
		case <-ctx.Done():
			return
		case resourcesInfo := <-resourceChan:
			c.processEachResource(resourcesInfo)
		}
	}
}

func (c *Transformer) processEachResource(results fetching.ResourcesInfo) {
	for _, resource := range results.Resources {
		if err := c.createBeatEvents(resource, results.CycleMetadata); err != nil {
			c.log.Errorf("Failed to create beat events for Cycle ID: %v, error: %v",
				results.CycleMetadata.CycleId, err)
		}
	}
}

func (c *Transformer) createBeatEvents(fetchedResource fetching.Resource, cycleMetadata fetching.CycleMetadata) error {
	resMetadata := fetchedResource.GetMetadata()
	resMetadata.ID = c.commonData.GetResourceId(resMetadata)
	fetcherResult := fetching.Result{Type: resMetadata.Type, Resource: fetchedResource.GetData()}

	result, err := c.eval.Decision(c.context, fetcherResult)

	if err != nil {
		c.log.Errorf("Error running the policy: %v", err)
		return err
	}

	c.log.Debugf("Eval decision for input: %v -- %v", fetcherResult, result)

	findings, err := c.eval.Decode(result)
	if err != nil {
		return err
	}

	c.log.Debugf("Created %d findings for input: %v", len(findings), fetcherResult)

	timestamp := time.Now()
	resource := fetching.ResourceFields{
		ResourceMetadata: resMetadata,
		Raw:              fetcherResult.Resource,
	}

	for _, finding := range findings {
		event := beat.Event{
			Meta:      c.eventMetadata,
			Timestamp: timestamp,
			Fields: common.MapStr{
				"resource":    resource,
				"resource_id": resMetadata.ID,   // Deprecated - kept for BC
				"type":        resMetadata.Type, // Deprecated - kept for BC
				"cycle_id":    cycleMetadata.CycleId,
				"result":      finding.Result,
				"rule":        finding.Rule,
			},
		}

		c.events.Data <- event
	}

	return nil
}
