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

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

const (
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
	var eventsToSend []beat.Event
	flushTicker := time.NewTicker(p.interval)
	for {
		select {
		case <-ctx.Done():
			p.log.Warnf("Publisher context is done: %v", ctx.Err())
			p.publish(&eventsToSend)
			return

		// Flush events to ES after a pre-defined interval, meant to clean residuals after a cycle is finished.
		case <-flushTicker.C:
			if len(eventsToSend) == 0 {
				continue
			}

			p.log.Infof("Publisher time interval reached")
			p.publish(&eventsToSend)

		// Flush events to ES when reaching a certain threshold
		case event, ok := <-ch:
			if !ok {
				p.log.Warn("Publisher channel is closed")
				p.publish(&eventsToSend)
				return
			}

			eventsToSend = append(eventsToSend, event...)
			if len(eventsToSend) < p.threshold {
				continue
			}

			p.log.Infof("Publisher buffer threshold:%d reached", p.threshold)
			p.publish(&eventsToSend)
		}
	}
}

func (p *Publisher) publish(events *[]beat.Event) {
	if len(*events) == 0 {
		return
	}

	p.log.With(ecsEventActionField, ecsEventActionValue, ecsEventCountField, len(*events)).
		Infof("Publishing %d events to elasticsearch", len(*events))
	p.client.PublishAll(*events)
	*events = nil
}
