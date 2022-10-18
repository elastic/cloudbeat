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

package add_orchestrator_metadata

import (
	"fmt"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/processors"
	jsprocessor "github.com/elastic/beats/v7/libbeat/processors/script/javascript/module/processor"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
)

const (
	oldProcessorName = "add_cluster_id"
	processorName    = "add_orchestrator_metadata"
	clusterNameKey   = "orchestrator.cluster.name"
	ClusterIdKey     = "cluster_id"
)

func init() {
	// This is for backward compatibility - it needs to be removed once we are 9.0 and above
	processors.RegisterPlugin(oldProcessorName, New)
	jsprocessor.RegisterPlugin("AddClusterID", New)

	processors.RegisterPlugin(processorName, New)
	jsprocessor.RegisterPlugin("AddOrchestratorMetadata", New)
}

type processor struct {
	config config
	helper ClusterHelper
	logger *logp.Logger
}

// New constructs a new orchestrator metadata processor.
func New(cfg *agentconfig.C) (processors.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, makeErrConfigUnpack(err)
	}

	client, err := kubernetes.GetKubernetesClient("", kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, err
	}

	logger := logp.NewLogger(processorName)
	clusterMetadataProvider, err := newClusterMetadataProvider(client, cfg, logger)

	if err != nil {
		return nil, err
	}
	p := &processor{
		config,
		clusterMetadataProvider,
		logger,
	}

	return p, nil
}

// Run enriches the given event with an ID
func (p *processor) Run(event *beat.Event) (*beat.Event, error) {
	clusterMetaData := p.helper.GetClusterMetadata()

	if _, err := event.PutValue(ClusterIdKey, clusterMetaData.clusterId); err != nil {
		return nil, makeErrComputeID(err)
	}

	clusterName := clusterMetaData.clusterName
	if clusterName != "" {
		_, err := event.PutValue(clusterNameKey, clusterName)
		if err != nil {
			return nil, fmt.Errorf("failed to add cluster name to object: %v", err)
		}
	}

	return event, nil
}

func (p *processor) String() string {
	return fmt.Sprintf("%v=", processorName)
}
