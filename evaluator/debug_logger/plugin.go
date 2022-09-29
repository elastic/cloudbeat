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

package dlogger

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/open-policy-agent/opa/plugins"
	"github.com/open-policy-agent/opa/plugins/logs"
	"github.com/open-policy-agent/opa/util"
)

const PluginName = "debug_decision_logs"

type Config struct {
}

type Plugin struct {
	manager *plugins.Manager
	mtx     sync.Mutex
	config  Config
}

func (p *Plugin) Start(ctx context.Context) error {
	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateOK})
	return nil
}

func (p *Plugin) Stop(ctx context.Context) {
	p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})
}

func (p *Plugin) Reconfigure(ctx context.Context, config interface{}) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.config = config.(Config)
}

// Log is called by the decision logger when a record (event) should be emitted. The logs.EventV1 fields
// map 1:1 to those described in https://www.openpolicyagent.org/docs/latest/management-decision-logs
func (p *Plugin) Log(ctx context.Context, event logs.EventV1) error {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	eventBuf, err := json.Marshal(&event)
	if err != nil {
		p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateErr})
		return nil
	}
	fields := map[string]interface{}{}
	err = util.UnmarshalJSON(eventBuf, &fields)
	if err != nil {
		p.manager.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateErr})
		return nil
	}
	p.manager.ConsoleLogger().WithFields(fields).WithFields(map[string]interface{}{
		"type": "openpolicyagent.org/decision_logs",
	}).Debug("Decision Log")
	return nil
}
