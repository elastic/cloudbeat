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

package fetchers

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"

	"github.com/elastic/beats/v7/x-pack/osquerybeat/ext/osquery-extension/pkg/proc"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	// CMDArgumentMatcher is a regex pattern that should match a process argument and its value
	// Expects format as the following: --<key><delimiter><value>.
	// For example: --config=a.json
	// The regex supports two delimiters "=" and ""
	CMDArgumentMatcher  = "\\b%s[\\s=]\\/?(\\S+)"
	ProcessResourceType = "process"
	ProcessSubType      = "process"
)

type ProcessResource struct {
	PID          string        `json:"pid"`
	Cmd          string        `json:"command"`
	Stat         proc.ProcStat `json:"stat"`
	ExternalData mapstr.M      `json:"external_data"`
}

type ProcessesFetcher struct {
	log        *logp.Logger
	cfg        ProcessFetcherConfig
	Fs         fs.FS
	resourceCh chan fetching.ResourceInfo
}

type ProcessInputConfiguration struct {
	ConfigFileArguments []string `config:"config-file-arguments"`
}

type ProcessesConfigMap map[string]ProcessInputConfiguration

type ProcessFetcherConfig struct {
	fetching.BaseFetcherConfig
	Directory         string             `config:"directory"` // parent directory of target procfs
	RequiredProcesses ProcessesConfigMap `config:"processes"`
}

func (f *ProcessesFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting ProcessesFetcher.Fetch")

	pids, err := proc.ListFS(f.Fs)
	if err != nil {
		return err
	}

	// If errors occur during read, then return what we have till now
	// without reporting errors.
	for _, p := range pids {
		stat, err := proc.ReadStatFS(f.Fs, p)
		if err != nil {
			return err
		}
		processConfig, isProcessRequired := f.cfg.RequiredProcesses[stat.Name]
		if !isProcessRequired {
			continue
		}

		fetchedResource, err := f.fetchProcessData(stat, processConfig, p)
		if err != nil {
			f.log.Error(err)
			continue
		}
		f.resourceCh <- fetching.ResourceInfo{Resource: fetchedResource, CycleMetadata: cMetadata}
	}

	return nil
}

func (f *ProcessesFetcher) fetchProcessData(procStat proc.ProcStat, processConf ProcessInputConfiguration, processId string) (fetching.Resource, error) {
	cmd, err := proc.ReadCmdLineFS(f.Fs, processId)
	if err != nil {
		return nil, err
	}
	configMap := f.getProcessConfigurationFile(processConf, cmd, procStat.Name)

	return ProcessResource{PID: processId, Cmd: cmd, Stat: procStat, ExternalData: configMap}, nil
}

// getProcessConfigurationFile - reads the configuration file associated with a process.
// As an input this function receives a ProcessInputConfiguration that contains ConfigFileArguments, a string array that represents some process flags
// The function extracts the configuration file associated with each flag and returns it.
func (f *ProcessesFetcher) getProcessConfigurationFile(processConfig ProcessInputConfiguration, cmd string, processName string) map[string]interface{} {
	configMap := make(map[string]interface{})
	for _, argument := range processConfig.ConfigFileArguments {
		// The regex extracts the cmd line flag(argument) value
		regex := fmt.Sprintf(CMDArgumentMatcher, argument)
		matcher := regexp.MustCompile(regex)
		if !matcher.MatchString(cmd) {
			f.log.Infof("Couldn't find a configuration file associated with flag %s for process %s from cmd", argument, processName, cmd)
			continue
		}

		groupMatches := matcher.FindStringSubmatch(cmd)
		if len(groupMatches) < 2 {
			f.log.Errorf("Couldn't find a configuration file associated with flag %s for process %s", argument, processName)
			continue
		}
		argValue := matcher.FindStringSubmatch(cmd)[1]
		f.log.Infof("Using %s as a configuration file for process %s", argValue, processName)

		data, err := fs.ReadFile(f.Fs, argValue)
		if err != nil {
			f.log.Errorf("Failed to read file configuration for process %s, error - %+v", processName, err)
			continue
		}
		configFile, err := f.readConfigurationFile(argValue, data)
		if err != nil {
			f.log.Errorf("Failed to parse file configuration for process %s, error - %+v", processName, err)
			continue
		}
		configMap[argument] = configFile
	}
	return configMap
}

func (f *ProcessesFetcher) readConfigurationFile(path string, data []byte) (interface{}, error) {
	ext := filepath.Ext(path)
	var output interface{}

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &output); err != nil {
			return nil, err
		}
	case ".yaml":
		if err := yaml.Unmarshal(data, &output); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%s type is not supported", ext)
	}
	return output, nil
}

func (f *ProcessesFetcher) Stop() {
}

func (res ProcessResource) GetData() interface{} {
	return res
}

func (res ProcessResource) GetMetadata() fetching.ResourceMetadata {
	return fetching.ResourceMetadata{
		ID:      res.PID + res.Stat.StartTime,
		Type:    ProcessResourceType,
		SubType: ProcessSubType,
		Name:    res.Stat.Name,
	}
}
