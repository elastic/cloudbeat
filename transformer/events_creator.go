package transformer

import (
	"context"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/beat"
	libevents "github.com/elastic/beats/v7/libbeat/beat/events"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources"
	"github.com/elastic/cloudbeat/resources/fetchers"
	"github.com/mitchellh/mapstructure"
	"time"
)

type CB func(ctx context.Context, input interface{}) (interface{}, error)

type Transformer struct {
	context       context.Context
	callback      CB
	eventMetadata common.MapStr
	events        []beat.Event
}

func NewTransformer(ctx context.Context, cb CB, index string) Transformer {
	eventMetadata := common.MapStr{libevents.FieldMetaIndex: index}
	events := make([]beat.Event, 0)

	return Transformer{
		context:       ctx,
		callback:      cb,
		eventMetadata: eventMetadata,
		events:        events,
	}
}

func (c *Transformer) ProcessAggregatedResources(resources resources.ResourceMap, metadata CycleMetadata) []beat.Event {
	c.events = make([]beat.Event, 0)
	for fetcherType, fetcherResults := range resources {
		c.processEachResource(fetcherResults, ResourceTypeMetadata{CycleMetadata: metadata, Type: fetcherType})
	}

	return c.events
}

func (c *Transformer) processEachResource(results []fetchers.PolicyResource, metadata ResourceTypeMetadata) {
	for _, result := range results {
		resMetadata := ResourceMetadata{ResourceTypeMetadata: metadata, ResourceId: result.GetID()}
		if err := c.createBeatEvents(result, resMetadata); err != nil {
			fmt.Errorf("failed to create beat events for, %v, Error: %v", metadata, err)

		}
	}
}

func (c *Transformer) createBeatEvents(policyResource fetchers.PolicyResource, metadata ResourceMetadata) error {
	fetcherResult := fetchers.FetcherResult{Type: metadata.Type, Resource: policyResource.GetData()}
	result, err := c.callback(c.context, fetcherResult)

	if err != nil {
		logp.Error(fmt.Errorf("error running the policy: %w", err))
		return err
	}

	findings, err := parseResult(result)
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

func parseResult(result interface{}) ([]Finding, error) {
	var opaResult RuleResult
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &opaResult})
	if err != nil {
		return nil, err
	}

	err = decoder.Decode(result)
	return opaResult.Findings, err
}
