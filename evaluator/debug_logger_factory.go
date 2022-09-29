package evaluator

import (
	"github.com/open-policy-agent/opa/plugins"
	"github.com/open-policy-agent/opa/util"
)

type DebugLoggerFactory struct{}

func (DebugLoggerFactory) New(m *plugins.Manager, config interface{}) plugins.Plugin {

	m.UpdatePluginStatus(PluginName, &plugins.Status{State: plugins.StateNotReady})

	return &DebugLogger{
		manager: m,
		config:  config.(Config),
	}
}

func (DebugLoggerFactory) Validate(_ *plugins.Manager, config []byte) (interface{}, error) {
	parsedConfig := Config{}
	return parsedConfig, util.Unmarshal(config, &parsedConfig)
}
