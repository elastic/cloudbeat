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
	fetcherResult := fetching.FetcherResult{Type: metadata.Type, Resource: fetchedResource.GetData()}
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
	for _, finding := range findings {
		event := beat.Event{
			Meta:      c.eventMetadata,
			Timestamp: timestamp,
			Fields: common.MapStr{
				"resource_id": metadata.ResourceId,
				"type":        metadata.Type,
				"cycle_id":    metadata.CycleId,
				"result":      finding.Result,
				"resource":    fetcherResult.Resource,
				"rule":        finding.Rule,
			},
		}

		c.events = append(c.events, event)
	}
	return nil
}
