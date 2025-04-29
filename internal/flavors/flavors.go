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
	"github.com/elastic/beats/v7/libbeat/management/status"

	"github.com/elastic/cloudbeat/internal/config"
	"github.com/elastic/cloudbeat/internal/infra/clog"
	_ "github.com/elastic/cloudbeat/internal/processor" // Add cloudbeat default processors.
)

const (
	flushInterval   = 10 * time.Second
	eventsThreshold = 75
)

// flavorBase configuration.
type flavorBase struct {
	ctx          context.Context //nolint:containedctx
	cancel       context.CancelFunc
	config       *config.Config
	client       beat.Client
	log          *clog.Logger
	publisher    *Publisher
	errorHandler ErrorHandler
}

type ErrorHandler interface {
	ErrorPublisher
	Start(ctx context.Context)
	Stop()
}

type ErrorProcessor interface {
	Process(status.StatusReporter, error)
}

type ErrorPublisher interface {
	Publish(ctx context.Context, err error)
}
