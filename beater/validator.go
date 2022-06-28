package beater

import (
	"fmt"

	"github.com/elastic/cloudbeat/config"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
)

type validator struct {
}

func (v *validator) Validate(cfg *agentconfig.C) error {
	c, err := config.New(cfg)
	if err != nil {
		return fmt.Errorf("Could not parse reconfiguration %v, skipping with error: %v", cfg.FlattenedKeys(), err)
	}

	if len(c.Streams) == 0 {
		return fmt.Errorf("No streams received in reconfiguration %v", cfg.FlattenedKeys())
	}

	if c.Streams[0].DataYaml == nil {
		return fmt.Errorf("data_yaml not present in reconfiguration %v", cfg.FlattenedKeys())
	}

	return nil
}
