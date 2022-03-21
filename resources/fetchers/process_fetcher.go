package fetchers

import (
	"context"
	"errors"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/x-pack/osquerybeat/ext/osquery-extension/pkg/proc"
	"github.com/elastic/cloudbeat/config"
	"github.com/elastic/cloudbeat/resources/fetching"
	"io/fs"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"path/filepath"
	"regexp"
)

const (
	CMDArgumentMatcher = "\\b%s=\\/?(\\S+)"
)

type ProcessResource struct {
	PID          string        `json:"pid"`
	Cmd          string        `json:"command"`
	Stat         proc.ProcStat `json:"stat"`
	ExternalData common.MapStr `json:"external_data"`
}

type ProcessesFetcher struct {
	cfg ProcessFetcherConfig
}

type ProcessFetcherConfig struct {
	fetching.BaseFetcherConfig
	Directory         string `config:"directory"` // parent directory of target procfs
	Fs                fs.FS
	RequiredProcesses config.ProcessesConfigMap `config:"required_processes"`
}

func (f *ProcessesFetcher) Fetch(ctx context.Context) ([]fetching.Resource, error) {
	pids, err := proc.ListFS(f.cfg.Fs)
	if err != nil {
		return nil, err
	}
	ret := make([]fetching.Resource, 0)

	// If errors occur during read, then return what we have till now
	// without reporting errors.
	for _, p := range pids {
		stat, err := proc.ReadStatFS(f.cfg.Fs, p)
		if err != nil {
			return nil, err
		}
		processConfig, isProcessRequired := f.cfg.RequiredProcesses[stat.Name]
		if !isProcessRequired {
			continue
		}

		fetchedResource, err := f.fetchProcessData(stat, processConfig, p)
		if err != nil {
			logp.Error(fmt.Errorf("%+v", err))
			continue
		}
		ret = append(ret, fetchedResource)
	}

	return ret, nil
}

func (f *ProcessesFetcher) fetchProcessData(procStat proc.ProcStat, processConf config.ProcessInputConfiguration, processId string) (fetching.Resource, error) {
	cmd, err := proc.ReadCmdLineFS(f.cfg.Fs, processId)
	if err != nil {
		return nil, err
	}
	configMap := f.getProcessConfigurationFile(processConf, cmd, procStat.Name)

	return ProcessResource{PID: processId, Cmd: cmd, Stat: procStat, ExternalData: configMap}, nil
}

//getProcessConfigurationFile - reads the configuration file associated with a process.
// As an input this function receives a ProcessInputConfiguration that contains CommandArguments, a string array that represents some process flags
// The function extracts the configuration file associated with each flag and returns it.
func (f *ProcessesFetcher) getProcessConfigurationFile(processConfig config.ProcessInputConfiguration, cmd string, processName string) map[string]interface{} {
	configMap := make(map[string]interface{}, 0)
	for _, argument := range processConfig.CommandArguments {
		// The regex extracts the cmd line flag(argument) value
		regex := fmt.Sprintf(CMDArgumentMatcher, argument)
		matcher := regexp.MustCompile(regex)
		if !matcher.MatchString(cmd) {
			logp.Error(fmt.Errorf("failed to find argument %s in process %s", argument, processName))
			continue
		}
		argValue := matcher.FindStringSubmatch(cmd)[1]
		data, err := fs.ReadFile(f.cfg.Fs, argValue)
		if err != nil {
			logp.Error(fmt.Errorf("failed to read file configuration for process %s, error - %+v", processName, err))
			continue
		}
		configFile, err := f.readConfigurationFile(argValue, data)
		if err != nil {
			logp.Error(fmt.Errorf("failed to parse file configuration for process %s, error - %+v", processName, err))
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
		return nil, errors.New("can't parse data")
	}
	return output, nil
}

func (f *ProcessesFetcher) Stop() {
}

func (res ProcessResource) GetID() (string, error) {
	return res.PID, nil
}

func (res ProcessResource) GetData() interface{} {
	return res
}
