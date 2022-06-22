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

package add_cluster_id

import (
	"fmt"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/processors"
	jsprocessor "github.com/elastic/beats/v7/libbeat/processors/script/javascript/module/processor"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/config"
)

func init() {
	processors.RegisterPlugin("add_cluster_id", New)
	jsprocessor.RegisterPlugin("AddClusterID", New)
}

const processorName = "add_cluster_id"

type addClusterID struct {
	config procConfig
	helper ClusterHelper
}

// New constructs a new Add ID processor.
func New(cfg *config.C) (processors.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, makeErrConfigUnpack(err)
	}

	client, err := kubernetes.GetKubernetesClient("", kubernetes.KubeClientOptions{})
	if err != nil {
		return nil, err
	}

	helper, err := newClusterHelper(client)
	if err != nil {
		return nil, err
	}
	p := &addClusterID{
		config,
		helper,
	}

	return p, nil
}

// Run enriches the given event with an ID
func (p *addClusterID) Run(event *beat.Event) (*beat.Event, error) {
	clusterId := p.helper.ClusterId()

	if _, err := event.PutValue(p.config.TargetField, clusterId); err != nil {
		return nil, makeErrComputeID(err)
	}

	return event, nil
}

func (p *addClusterID) String() string {
	return fmt.Sprintf("%v=[target_field=[%v]]", processorName, p.config.TargetField)
}
