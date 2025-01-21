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
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/beats/v7/x-pack/osquerybeat/ext/osquery-extension/pkg/proc"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/utils/clog"
)

const (
	// CMDArgumentMatcher is a regex pattern that should match a process argument and its value
	// Expects format as the following: --<key><delimiter><value>.
	// For example: --config=a.json
	// The regex supports two delimiters "=" and ""
	CMDArgumentMatcher  = "\\b%s[\\s=]\\/?(\\S+)"
	ProcessResourceType = "process"
	ProcessSubType      = "process"
	directory           = "/hostfs"
	userHz              = 100
)

type EvalProcResource struct {
	PID          string        `json:"pid"`
	Cmd          string        `json:"command"`
	Stat         proc.ProcStat `json:"stat"`
	ExternalData mapstr.M      `json:"external_data"`
}

// ProcCommonData According to https://www.elastic.co/guide/en/ecs/current/ecs-process.html
type ProcCommonData struct {
	// Parent process.
	Parent *ProcCommonData `json:"parent,omitempty"`

	// Process id.
	PID int64 `json:"pid,omitempty"`

	// Process name.
	// Sometimes called program name or similar.
	Name string `json:"name,omitempty"`

	// Identifier of the group of processes the process belongs to.
	PGID int64 `json:"pgid,omitempty"`

	// Full command line that started the process, including the absolute path
	// to the executable, and all arguments.
	// Some arguments may be filtered to protect sensitive information.
	CommandLine string `json:"command_line,omitempty"`

	// Array of process arguments, starting with the absolute path to the
	// executable.
	// May be filtered to protect sensitive information.
	Args []string `json:"args,omitempty"`

	// Length of the process.args array.
	// This field can be useful for querying or performing bucket analysis on
	// how many arguments were provided to start a process. More arguments may
	// be an indication of suspicious activity.
	ArgsCount int64 `json:"args_count,omitempty"`

	// Process title.
	// The proctitle, sometimes the same as process name. Can also be
	// different: for example a browser setting its title to the web page
	// currently opened.
	Title string `json:"title,omitempty"`

	// The time the process started.
	Start time.Time `json:"start"`

	// Seconds the process has been up.
	Uptime int64 `json:"uptime,omitempty"`
}

type ProcResource struct {
	EvalResource  EvalProcResource
	ElasticCommon ProcCommonData
}

type ProcessesFetcher struct {
	log        *clog.Logger
	Fs         fs.FS
	resourceCh chan fetching.ResourceInfo
	processes  ProcessesConfigMap
}

type ProcessInputConfiguration struct {
	ConfigFileArguments []string `config:"config-file-arguments"`
}

type ProcessesConfigMap map[string]ProcessInputConfiguration

func NewProcessFetcher(log *clog.Logger, ch chan fetching.ResourceInfo, processes ProcessesConfigMap) *ProcessesFetcher {
	return &ProcessesFetcher{
		log:        log,
		Fs:         os.DirFS(directory),
		resourceCh: ch,
		processes:  processes,
	}
}

func (f *ProcessesFetcher) Fetch(_ context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Debug("Starting ProcessesFetcher.Fetch")

	pids, err := proc.ListFS(f.Fs)
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	// If errors occur during read, then return what we have till now
	// without reporting errors.
	for _, p := range pids {
		stat, err := proc.ReadStatFS(f.Fs, p)
		if err != nil {
			f.log.Errorf("error while reading /proc/<pid>/stat for process %s: %s", p, err.Error())
			continue
		}

		// Get the full command line name and not the /proc/pid/status one which might be silently truncated.
		cmd, err := proc.ReadCmdLineFS(f.Fs, p)
		if err != nil {
			f.log.Error("error while reading /proc/<pid>/cmdline for process %s: %s", p, err.Error())
			continue
		}
		name := extractCommandName(cmd)

		processConfig, isProcessRequired := f.processes[name]
		if !isProcessRequired {
			continue
		}

		fetchedResource := f.fetchProcessData(stat, processConfig, p, cmd)
		f.resourceCh <- fetching.ResourceInfo{Resource: fetchedResource, CycleMetadata: cycleMetadata}
	}

	return nil
}

func (f *ProcessesFetcher) fetchProcessData(procStat proc.ProcStat, processConf ProcessInputConfiguration, processId string, cmd string) fetching.Resource {
	configMap := f.getProcessConfigurationFile(processConf, cmd, procStat.Name)
	evalRes := EvalProcResource{PID: processId, Cmd: cmd, Stat: procStat, ExternalData: configMap}
	procCd := f.createProcCommonData(procStat, cmd, processId)
	return ProcResource{EvalResource: evalRes, ElasticCommon: procCd}
}

func (f *ProcessesFetcher) createProcCommonData(stat proc.ProcStat, cmd string, pid string) ProcCommonData {
	processID, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		f.log.Errorf("Couldn't parse PID, pid: %s", pid)
	}

	startTime, err := strconv.ParseUint(stat.StartTime, 10, 64)
	if err != nil {
		f.log.Errorf("Couldn't parse stat.StartTime, startTime: %s", stat.StartTime)
	}

	pgid, err := strconv.ParseInt(stat.Group, 10, 64)
	if err != nil {
		f.log.Errorf("Couldn't parse stat.Group, Group: %s, Error: %v", stat.Group, err)
	}

	ppid, err := strconv.ParseInt(stat.Parent, 10, 64)
	if err != nil {
		f.log.Errorf("Couldn't parse stat.Parent, Parent: %s, Error: %v", stat.Parent, err)
	}

	sysUptime, err := proc.ReadUptimeFS(f.Fs)
	if err != nil {
		f.log.Error("couldn't read system boot time", err)
	}
	uptimeDate := time.Now().Add(-time.Duration(sysUptime) * time.Second)

	args := strings.Split(cmd, " ")
	start := uptimeDate.Add(ticksToDuration(startTime))
	return ProcCommonData{
		Parent:      &ProcCommonData{PID: ppid},
		PID:         processID,
		Name:        stat.Name,
		PGID:        pgid,
		CommandLine: cmd,
		Args:        args,
		ArgsCount:   int64(len(args)),
		Title:       stat.Name,
		Start:       start,
		Uptime:      int64(time.Since(start).Seconds()),
	}
}

// getProcessConfigurationFile - reads the configuration file associated with a process.
// As an input this function receives a ProcessInputConfiguration that contains ConfigFileArguments, a string array that represents some process flags
// The function extracts the configuration file associated with each flag and returns it.
func (f *ProcessesFetcher) getProcessConfigurationFile(processConfig ProcessInputConfiguration, cmd string, processName string) map[string]any {
	configMap := make(map[string]any)
	for _, argument := range processConfig.ConfigFileArguments {
		// The regex extracts the cmd line flag(argument) value
		regex := fmt.Sprintf(CMDArgumentMatcher, argument)
		matcher := regexp.MustCompile(regex)
		if !matcher.MatchString(cmd) {
			f.log.Infof("Couldn't find a configuration file associated with flag %s for process %s from cmd %s", argument, processName, cmd)
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

func (f *ProcessesFetcher) readConfigurationFile(path string, data []byte) (any, error) {
	ext := filepath.Ext(path)
	var output any

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

func (res ProcResource) GetData() any {
	return res.EvalResource
}

func (res ProcResource) GetIds() []string {
	return nil
}

func (res ProcResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      res.EvalResource.PID + res.EvalResource.Stat.StartTime,
		Type:    ProcessResourceType,
		SubType: ProcessSubType,
		Name:    res.EvalResource.Stat.Name,
	}, nil
}

func (res ProcResource) GetElasticCommonData() (map[string]any, error) {
	m := map[string]any{}
	m["process.parent.pid"] = res.ElasticCommon.Parent.PID
	m["process.pid"] = res.ElasticCommon.PID
	m["process.name"] = res.ElasticCommon.Name
	m["process.pgid"] = res.ElasticCommon.PGID
	m["process.command_line"] = res.ElasticCommon.CommandLine
	m["process.args"] = res.ElasticCommon.Args
	m["process.args_count"] = res.ElasticCommon.ArgsCount
	m["process.title"] = res.ElasticCommon.Title
	m["process.start"] = res.ElasticCommon.Start
	m["process.uptime"] = res.ElasticCommon.Uptime

	return m, nil
}

// Supported only in Linux
func ticksToDuration(ticks uint64) time.Duration {
	seconds := float64(ticks) / float64(userHz) * float64(time.Second)
	return time.Duration(int64(seconds))
}

func extractCommandName(cmdline string) string {
	// remove command line arguments by finding the first space.
	// <root>/proc/pid/cmdline separates the strings with null bytes ('\0'),
	// but proc.ReadCmdLineFS replaces them with space.
	i := strings.IndexByte(cmdline, ' ')
	if i > -1 {
		cmdline = cmdline[:i]
	}

	// remove the path (if exists) and return the process' executable file name.
	_, file := path.Split(cmdline)

	return file
}
