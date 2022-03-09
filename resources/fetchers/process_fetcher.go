package fetchers

import (
	"context"
	"errors"
	"fmt"
	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/x-pack/osquerybeat/ext/osquery-extension/pkg/proc"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
	"regexp"
)

const (
	ProcessType        = "process"
	CMDArgumentMatcher = "\\b%s=(\\S+)"
)

type ProcessResource struct {
	PID    string                 `json:"pid"`
	Cmd    string                 `json:"command"`
	Stat   proc.ProcStat          `json:"stat"`
	Config map[string]interface{} `json:"config"`
}

type ProcessesFetcher struct {
	cfg ProcessFetcherConfig
}

type ProcessFetcherConfig struct {
	BaseFetcherConfig
	Directory         string                               `config:"directory"` // parent directory of target procfs
	RequiredProcesses map[string]ProcessInputConfiguration `config:"required_processes"`
}

type ProcessInputConfiguration struct {
	RequiredFileInputArguments []string `config:"required-file-arguments"`
}

func NewProcessesFetcher(cfg ProcessFetcherConfig) Fetcher {
	return &ProcessesFetcher{
		cfg: cfg,
	}
}

func (f *ProcessesFetcher) Fetch(ctx context.Context) ([]FetchedResource, error) {
	pids, err := proc.List(f.cfg.Directory)
	if err != nil {
		return nil, err
	}

	ret := make([]FetchedResource, 0)

	// If errors occur during read, then return what we have till now
	// without reporting errors.
	for _, p := range pids {
		stat, err := proc.ReadStat(f.cfg.Directory, p)
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

func (f *ProcessesFetcher) fetchProcessData(procStat proc.ProcStat, process ProcessInputConfiguration, processId string) (FetchedResource, error) {
	cmd, err := proc.ReadCmdLine(f.cfg.Directory, processId)
	if err != nil {
		return nil, err
	}

	configMap := f.getProcessFileConfiguration(process, cmd, procStat.Name)

	return ProcessResource{PID: processId, Cmd: cmd, Stat: procStat, Config: configMap}, nil
}

func (f *ProcessesFetcher) getProcessFileConfiguration(process ProcessInputConfiguration, cmd string, processName string) map[string]interface{} {
	configMap := make(map[string]interface{}, 0)
	for _, config := range process.RequiredFileInputArguments {
		regex := fmt.Sprintf(CMDArgumentMatcher, config)
		matcher := regexp.MustCompile(regex)
		if !matcher.MatchString(cmd) {
			logp.Error(fmt.Errorf("failed to find argument %s to process %s", config, processName))
			continue
		}
		configPath := filepath.Join(f.cfg.Directory, matcher.FindStringSubmatch(cmd)[1])
		data, err := os.ReadFile(configPath)
		if err != nil {
			logp.Error(fmt.Errorf("failed to read file configuration for process %s, error - %+v", processName, err))
			continue
		}
		configFile, err := f.readConfigurationFile(configPath, data)
		if err != nil {
			logp.Error(fmt.Errorf("failed to parse file configuration for process %s, error - %+v", processName, err))
			continue
		}
		configMap[config] = configFile
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
