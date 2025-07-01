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

package flavors

import (
	"context"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/elastic/cloudbeat/internal/infra/clog"
	"github.com/elastic/cloudbeat/internal/infra/observability"
)

const (
	scopeName = "github.com/elastic/cloudbeat/internal/flavors"

	ecsEventActionField = "event.action"
	ecsEventActionValue = "publish-events"
	ecsEventCountField  = "event.Count"
)

type client interface {
	PublishAll([]beat.Event)
}

type Publisher struct {
	log       *clog.Logger
	interval  time.Duration
	threshold int
	client    client
}

func NewPublisher(log *clog.Logger, interval time.Duration, threshold int, client client) *Publisher {
	return &Publisher{
		log:       log,
		interval:  interval,
		threshold: threshold,
		client:    client,
	}
}

func (p *Publisher) HandleEvents(ctx context.Context, ch <-chan []beat.Event) {
	count, err := observability.MeterFromContext(ctx, scopeName).Int64Counter("cloudbeat.events.published")
	if err != nil {
		panic("failed to create events published counter: " + err.Error()) // TODO: log
	}

	var eventsToSend []beat.Event
	flushTicker := time.NewTicker(p.interval)
	for {
		select {
		case <-ctx.Done():
			p.log.Warnf("Publisher context is done: %v", ctx.Err())
			p.publish(ctx, &eventsToSend, count)
			return

		// Flush events to ES after a pre-defined interval, meant to clean residuals after a cycle is finished.
		case <-flushTicker.C:
			if len(eventsToSend) == 0 {
				continue
			}

			p.log.Infof("Publisher time interval reached")
			p.publish(ctx, &eventsToSend, count)

		// Flush events to ES when reaching a certain threshold
		case event, ok := <-ch:
			if !ok {
				p.log.Warn("Publisher channel is closed")
				p.publish(ctx, &eventsToSend, count)
				return
			}

			eventsToSend = append(eventsToSend, event...)
			if len(eventsToSend) < p.threshold {
				continue
			}

			p.log.Infof("Publisher buffer threshold:%d reached", p.threshold)
			p.publish(ctx, &eventsToSend, count)
		}
	}
}

func (p *Publisher) publish(ctx context.Context, events *[]beat.Event, count metric.Int64Counter) {
	batchSize := len(*events)
	if batchSize == 0 {
		return
	}
	ctx, span := observability.StartSpan(
		ctx,
		scopeName,
		"Publish Events",
		trace.WithSpanKind(trace.SpanKindProducer),
	)
	defer span.End()

	p.log.
		WithSpanContext(span.SpanContext()).
		With(
			ecsEventActionField, ecsEventActionValue,
			ecsEventCountField, batchSize,
		).
		Infof("Publishing %d events to elasticsearch", batchSize)

	p.client.PublishAll(*events)
	count.Add(ctx, int64(batchSize))

	*events = nil
}
