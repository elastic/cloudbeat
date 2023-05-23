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

package cmd

import (
	"fmt"
	"github.com/elastic/cloudbeat/version"

	"github.com/elastic/cloudbeat/beater"
	"github.com/elastic/elastic-agent-client/v7/pkg/client"
	"github.com/elastic/elastic-agent-client/v7/pkg/proto"

	"github.com/elastic/beats/v7/libbeat/cmd"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"
	"github.com/elastic/beats/v7/libbeat/common/reload"

	_ "github.com/elastic/beats/v7/x-pack/libbeat/include"
	"github.com/elastic/beats/v7/x-pack/libbeat/management"
)

// Name of this beat
var Name = "cloudbeat"

// RootCmd to handle beats cli
var RootCmd = cmd.GenRootCmdWithSettings(beater.New, instance.Settings{Name: Name, Version: version.CloudbeatSemanticVersion()})

func cloudbeatCfg(rawIn *proto.UnitExpectedConfig, agentInfo *client.AgentInfo) ([]*reload.ConfigWithMeta, error) {
	modules, err := management.CreateInputsFromStreams(rawIn, "logs", agentInfo)
	if err != nil {
		return nil, fmt.Errorf("error creating input list from raw expected config: %w", err)
	}

	// format for the reloadable list needed by the cm.Reload() method
	configList, err := management.CreateReloadConfigFromInputs(modules)
	if err != nil {
		return nil, fmt.Errorf("error creating reloader config: %w", err)
	}

	return configList, nil
}

func init() {
	management.ConfigTransform.SetTransform(cloudbeatCfg)
}
