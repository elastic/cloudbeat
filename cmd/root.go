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
	"errors"
	"fmt"

	"github.com/elastic/beats/v7/libbeat/cmd"
	"github.com/elastic/beats/v7/libbeat/cmd/instance"
	"github.com/elastic/beats/v7/libbeat/common/reload"
	"github.com/elastic/beats/v7/libbeat/publisher/processing"
	_ "github.com/elastic/beats/v7/x-pack/libbeat/include" // Initialize x-pack components
	"github.com/elastic/beats/v7/x-pack/libbeat/management"
	"github.com/elastic/elastic-agent-client/v7/pkg/client"
	"github.com/elastic/elastic-agent-client/v7/pkg/proto"

	"github.com/elastic/cloudbeat/internal/beater"
	"github.com/elastic/cloudbeat/version"
)

// Name of this beat
var Name = "cloudbeat"

// RootCmd to handle beats cli
var RootCmd = cmd.GenRootCmdWithSettings(
	beater.New,
	instance.Settings{
		Name:    Name,
		Version: version.CloudbeatSemanticVersion(),
		// Supply our own processing pipeline. Same as processing.MakeDefaultBeatSupport, but without
		// `processing.WithHost`.
		Processing: processing.MakeDefaultSupport(true, nil, processing.WithECS, processing.WithAgentMeta()),
	},
)

func cloudbeatCfg(rawIn *proto.UnitExpectedConfig, agentInfo *client.AgentInfo) ([]*reload.ConfigWithMeta, error) {
	modules, err := management.CreateInputsFromStreams(rawIn, "logs", agentInfo)
	if err != nil {
		return nil, fmt.Errorf("error creating input list from raw expected config: %w", err)
	}

	config := rawIn.Source.AsMap()
	packagePolicyID, ok := config["package_policy_id"]
	if !ok {
		return nil, errors.New("'package_policy_id' element does not exist")
	}

	packagePolicyRevision, ok := config["revision"]
	if !ok {
		return nil, errors.New("'revision' element does not exist")
	}

	for i := range modules {
		modules[i]["package_policy_id"] = packagePolicyID
		modules[i]["revision"] = packagePolicyRevision
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
