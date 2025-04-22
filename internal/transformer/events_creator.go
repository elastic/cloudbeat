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
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/gofrs/uuid"
	"github.com/samber/lo"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/dataprovider"
	"github.com/elastic/cloudbeat/internal/evaluator"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/resources/fetching"
	gcpfetchers "github.com/elastic/cloudbeat/internal/resources/fetching/fetchers/gcp"
)

const (
	ecsCategoryConfiguration = "configuration"
	ecsKindState             = "state"
	ecsTypeInfo              = "info"
)

const (
	EcsOutcomeSuccess = "success"
	EcsOutcomeFailure = "failure"
	EcsOutcomeUnknown = "unknown"
)

type Transformer struct {
	log                   *clog.Logger
	index                 string
	idProvider            dataprovider.IdProvider
	benchmarkDataProvider dataprovider.CommonDataProvider
	commonDataProvider    dataprovider.ElasticCommonDataProvider
}

type ECSEvent struct {
	Category []string  `json:"category"`
	Created  time.Time `json:"created"`
	ID       string    `json:"id"`
	Kind     string    `json:"kind"`
	Sequence int64     `json:"sequence"`
	Outcome  string    `json:"outcome"`
	Type     []string  `json:"type"`
}

type Related struct {
	Entity []string `json:"entity,omitempty"`
}

func NewTransformer(log *clog.Logger, cfg *config.Config, bdp dataprovider.CommonDataProvider, cdp dataprovider.ElasticCommonDataProvider, idp dataprovider.IdProvider) *Transformer {
	return &Transformer{
		log:                   log,
		index:                 cfg.Datastream(),
		idProvider:            idp,
		benchmarkDataProvider: bdp,
		commonDataProvider:    cdp,
	}
}

func (t *Transformer) CreateBeatEvents(_ context.Context, eventData evaluator.EventData) ([]beat.Event, error) {
	if len(eventData.Findings) == 0 {
		return nil, nil
	}

	events := make([]beat.Event, 0)
	resMetadata, err := eventData.GetMetadata()
	if err != nil {
		return []beat.Event{}, fmt.Errorf("failed to get resource metadata: %v", err)
	}
	id := t.idProvider.GetId(resMetadata.Type, resMetadata.ID)
	t.log.Infof("resource of type %s with id %s got a new id %s", resMetadata.Type, resMetadata.ID, id)
	resMetadata.ID = id
	timestamp := time.Now().UTC()
	resource := fetching.ResourceFields{
		ResourceMetadata: resMetadata,
		Raw:              getPreferredRawValue(eventData),
	}

	related := Related{
		Entity: lo.Filter(eventData.GetIds(), func(item string, _ int) bool { return item != "" }),
	}

	globalEnricher := dataprovider.NewEnricher(t.commonDataProvider)

	for _, finding := range eventData.Findings {
		event := beat.Event{
			Meta:      mapstr.M{libevents.FieldMetaIndex: t.index},
			Timestamp: timestamp,

			Fields: mapstr.M{
				"event":    BuildECSEvent(eventData.CycleMetadata.Sequence, eventData.Metadata.CreatedAt, []string{ecsCategoryConfiguration}, getEcsOutcome(finding.Result.Evaluation)),
				"resource": resource,
				"result":   finding.Result,
				"rule":     finding.Rule,
				"related":  related,
				"message":  fmt.Sprintf("Rule %q: %s", finding.Rule.Name, finding.Result.Evaluation),
			},
		}

		err := t.benchmarkDataProvider.EnrichEvent(&event, resMetadata)
		if err != nil {
			return nil, fmt.Errorf("failed to enrich event with benchmark context: %v", err)
		}

		err = globalEnricher.EnrichEvent(&event)
		if err != nil {
			return nil, fmt.Errorf("failed to enrich event with global context: %v", err)
		}

		err = dataprovider.NewEnricher(eventData).EnrichEvent(&event)
		if err != nil {
			return nil, fmt.Errorf("failed to enrich event with resource context: %v", err)
		}

		events = append(events, event)
	}

	return events, nil
}

func BuildECSEvent(seq int64, created time.Time, categories []string, outcome string) ECSEvent {
	id, _ := uuid.NewV4() // zero value in case of an error is uuid.Nil
	return ECSEvent{
		Category: categories,
		Created:  created,
		ID:       id.String(),
		Kind:     ecsKindState,
		Sequence: seq,
		Outcome:  outcome,
		Type:     []string{ecsTypeInfo},
	}
}

func getEcsOutcome(evaluation string) string {
	switch evaluation {
	case "failed":
		return EcsOutcomeFailure
	case "passed":
		return EcsOutcomeSuccess
	default:
		return EcsOutcomeUnknown
	}
}

func getPreferredRawValue(event evaluator.EventData) any {
	switch event.ResourceInfo.Resource.(type) {
	// Skip 'raw' for GCP custom resources grouping real GCP resources
	case *gcpfetchers.GcpLoggingAsset,
		*gcpfetchers.GcpMonitoringAsset,
		*gcpfetchers.GcpPoliciesAsset,
		*gcpfetchers.GcpServiceUsageAsset:
		return nil
	default:
		return event.RuleResult.Resource
	}
}
