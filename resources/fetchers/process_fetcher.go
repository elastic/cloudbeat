package fetchers

import (
	"context"
	"errors"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/x-pack/osquerybeat/ext/osquery-extension/pkg/proc"
	"github.com/elastic/cloudbeat/config"
	"io/fs"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"path/filepath"
	"regexp"
)

const (
	ProcessType        = "process"
	CMDArgumentMatcher = "\\b%s=(\\S+)"
)

type ProcessResource struct {
	PID    string        `json:"pid"`
	Cmd    string        `json:"command"`
	Stat   proc.ProcStat `json:"stat"`
	Config common.MapStr `json:"config"`
}

type ProcessesFetcher struct {
	cfg ProcessFetcherConfig
}

type ProcessFetcherConfig struct {
	BaseFetcherConfig
	Fs                fs.FS                     `config:"required_processes"`
	RequiredProcesses config.ProcessesConfigMap `config:"required_processes"`
}

func NewProcessesFetcher(cfg ProcessFetcherConfig) Fetcher {
	return &ProcessesFetcher{
		cfg: cfg,
	}
}

func (f *ProcessesFetcher) Fetch(ctx context.Context) ([]FetchedResource, error) {
	pids, err := proc.ListFS(f.cfg.Fs)
	if err != nil {
		return nil, err
	}
	ret := make([]FetchedResource, 0)

	// If errors occur during read, then return what we have till now
	// without reporting errors.
	for _, p := range pids {
		stat, err := proc.ReadStatFS(f.cfg.Fs, p)
		if err != nil {
			return nil, err
		}
		processInput, isProcessRequired := f.cfg.RequiredProcesses[stat.Name]
		if !isProcessRequired {
			continue
		}

		fetchedResource, err := f.fetchProcessData(stat, processInput, p)
		if err != nil {
			logp.Error(fmt.Errorf("%+v", err))
			continue
		}
		ret = append(ret, fetchedResource)
	}

	return ret, nil
}

func (f *ProcessesFetcher) fetchProcessData(procStat proc.ProcStat, process config.ProcessInputConfiguration, processId string) (FetchedResource, error) {
	cmd, err := proc.ReadCmdLineFS(f.cfg.Fs, processId)
	if err != nil {
		return nil, err
	}

	configMap := f.getProcessConfigurationFile(process, cmd, procStat.Name)

	return ProcessResource{PID: processId, Cmd: cmd, Stat: procStat, Config: configMap}, nil
}

//getProcessConfigurationFile - This function meant for reading the configuration file associated with a process.
// As an input this function receives a ProcessInputConfiguration that contains CommandArguments, a string array that represents some process flags
// that are related to the process configuration.
// The function extracts the file path of each of the CommandArguments And returns the files associated with them.
func (f *ProcessesFetcher) getProcessConfigurationFile(processConfig config.ProcessInputConfiguration, cmd string, processName string) map[string]interface{} {
	configMap := make(map[string]interface{}, 0)
	for _, argument := range processConfig.CommandArguments {
		// The regex extract the flag value of argument out of the process cmd line
		regex := fmt.Sprintf(CMDArgumentMatcher, argument)
		matcher := regexp.MustCompile(regex)
		if !matcher.MatchString(cmd) {
			logp.Error(fmt.Errorf("failed to find argument %s to processConfig %s", argument, processName))
			continue
		}
		// Since the process is mounted we need to add the mounted directory as Prefix
		// It won't work if the config file directory wasn't mounted
		configPath := matcher.FindStringSubmatch(cmd)[1]
		data, err := fs.ReadFile(f.cfg.Fs, configPath)
		if err != nil {
			logp.Error(fmt.Errorf("failed to read file configuration for processConfig %s, error - %+v", processName, err))
			continue
		}
		configFile, err := f.readConfigurationFile(configPath, data)
		if err != nil {
			logp.Error(fmt.Errorf("failed to parse file configuration for processConfig %s, error - %+v", processName, err))
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
