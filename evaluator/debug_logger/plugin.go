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
