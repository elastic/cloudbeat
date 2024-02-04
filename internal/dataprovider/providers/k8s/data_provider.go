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

package k8s

import (
	"errors"

	"github.com/elastic/beats/v7/libbeat/beat"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
)

const (
	clusterNameField    = "orchestrator.cluster.name"
	clusterVersionField = "orchestrator.cluster.version"
	clusterIdField      = "orchestrator.cluster.id"
	orchestratorType    = "orchestrator.type"
	orchestratorName    = "kubernetes"
)

type DataProvider struct {
	cluster        string
	clusterID      string
	clusterVersion string
}

func New(options ...Option) DataProvider {
	kdp := DataProvider{}
	for _, opt := range options {
		opt(&kdp)
	}
	return kdp
}

func (k DataProvider) EnrichEvent(event *beat.Event, _ fetching.ResourceMetadata) error {
	return errors.Join(
		insertIfNotEmpty(clusterNameField, k.cluster, event),
		insertIfNotEmpty(clusterIdField, k.clusterID, event),
		insertIfNotEmpty(clusterVersionField, k.clusterVersion, event),
		insertIfNotEmpty(orchestratorType, orchestratorName, event),
	)
}

func insertIfNotEmpty(field string, value string, event *beat.Event) error {
	if value != "" {
		_, err := event.Fields.Put(field, value)
		return err
	}
	return nil
}
