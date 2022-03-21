package fetchers

import (
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/fs"
	"k8s.io/apimachinery/pkg/util/json"
	"testing"
	"testing/fstest"
)

const (
	statContent = `1167 (containerd-shim) S 1 1167 198 0 -1 1077952768 223005 9831 39 0 665 1329 8 10 20 0 12 0 76222 730476544 2268 18446744073709551615 1 1 0 0 0 0 1006249984 0 2143420159 0 0 0 17 2 0 0 0 0 0 0 0 0 0 0 0 0 0`
)

type TextProcessContext struct {
	Pid               string
	Name              string
	ConfigFileFlagKey string
	ConfigFilePath    string
}

type ProcessConfigTestStruct struct {
	A string
	B int
}

var status = `Name:   %s`
var cmdline = `/usr/bin/%s --kubeconfig=/etc/kubernetes/kubelet.conf --%s=%s`

func TestFetchWhenFlagExistsButNoFile(t *testing.T) {
	testProcess := TextProcessContext{
		Pid:               "3",
		Name:              "kubelet",
		ConfigFileFlagKey: "fetcherConfig",
		ConfigFilePath:    "test/path",
	}
	sysfs := createProcess(testProcess)

	fetcherConfig := ProcessFetcherConfig{
		BaseFetcherConfig: fetching.BaseFetcherConfig{},
		RequiredProcesses: map[string]ProcessInputConfiguration{
			"kubelet": {CommandArguments: []string{"fetcherConfig"}}},
		Fs: sysfs,
	}
	processesFetcher := &ProcessesFetcher{cfg: fetcherConfig}

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(fetchedResource), 1)

	processResource := fetchedResource[0].(ProcessResource)
	assert.Equal(t, processResource.PID, testProcess.Pid)
	assert.Equal(t, processResource.Stat.Name, "kubelet")
	assert.Contains(t, processResource.Cmd, "/usr/bin/kubelet")
}

func TestFetchWhenProcessDoesNotExist(t *testing.T) {
	testProcess := TextProcessContext{
		Pid:               "3",
		Name:              "kubelet",
		ConfigFileFlagKey: "fetcherConfig",
		ConfigFilePath:    "test/path",
	}
	fsys := createProcess(testProcess)

	fetcherConfig := ProcessFetcherConfig{
		BaseFetcherConfig: fetching.BaseFetcherConfig{},
		RequiredProcesses: map[string]ProcessInputConfiguration{
			"someProcess": {CommandArguments: []string{"fetcherConfig"}}},
		Fs: fsys,
	}
	processesFetcher := &ProcessesFetcher{cfg: fetcherConfig}

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(fetchedResource), 0)
}

func TestFetchWhenNoFlagRequired(t *testing.T) {
	testProcess := TextProcessContext{
		Pid:               "3",
		Name:              "kubelet",
		ConfigFileFlagKey: "fetcherConfig",
		ConfigFilePath:    "test/path",
	}
	fsys := createProcess(testProcess)

	fetcherConfig := ProcessFetcherConfig{
		BaseFetcherConfig: fetching.BaseFetcherConfig{},
		RequiredProcesses: map[string]ProcessInputConfiguration{
			"kubelet": {CommandArguments: []string{}}},
		Fs: fsys,
	}
	processesFetcher := &ProcessesFetcher{cfg: fetcherConfig}

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(fetchedResource), 1)

	processResource := fetchedResource[0].(ProcessResource)
	assert.Equal(t, processResource.PID, testProcess.Pid)
	assert.Equal(t, processResource.Stat.Name, "kubelet")
	assert.Contains(t, processResource.Cmd, "/usr/bin/kubelet")
}

func TestFetchWhenFlagExistsWithConfigFile(t *testing.T) {

	testCases := []struct {
		configFileName string
		marshal        func(in interface{}) (out []byte, err error)
		configType     string
	}{
		{"kubeletConfig.yaml", yaml.Marshal, "yaml"},
		{"kubeletConfig.json", json.Marshal, "json"},
	}

	for _, test := range testCases {
		configFlagKey := "fetcherConfig"
		// Creating a yaml file for the process fetcherConfig
		processConfig := ProcessConfigTestStruct{
			A: "A",
			B: 2,
		}
		configData, err := test.marshal(&processConfig)

		testProcess := TextProcessContext{
			Pid:               "3",
			Name:              "kubelet",
			ConfigFileFlagKey: configFlagKey,
			ConfigFilePath:    test.configFileName,
		}

		sysfs := createProcess(testProcess).(fstest.MapFS)
		sysfs[test.configFileName] = &fstest.MapFile{
			Data: []byte(configData),
		}

		fetcherConfig := ProcessFetcherConfig{
			BaseFetcherConfig: fetching.BaseFetcherConfig{},
			RequiredProcesses: map[string]ProcessInputConfiguration{
				"kubelet": {CommandArguments: []string{"fetcherConfig"}}},
			Fs: sysfs,
		}
		processesFetcher := &ProcessesFetcher{cfg: fetcherConfig}

		fetchedResource, err := processesFetcher.Fetch(context.TODO())
		assert.Nil(t, err)
		assert.Equal(t, len(fetchedResource), 1)

		processResource := fetchedResource[0].(ProcessResource)
		assert.Equal(t, processResource.PID, testProcess.Pid)
		assert.Equal(t, processResource.Stat.Name, "kubelet")
		assert.Contains(t, processResource.Cmd, "/usr/bin/kubelet")

		configResource := processResource.ExternalData[configFlagKey]
		var result ProcessConfigTestStruct
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &result})
		assert.Nil(t, err, "Could not decode process fetcherConfig result from %s type", test.configType)
		err = decoder.Decode(configResource)
		assert.Nil(t, err, "Could not decode process fetcherConfig result from file %s", test.configFileName)

		assert.Equal(t, result.A, processConfig.A)
		assert.Equal(t, result.B, processConfig.B)
	}
}

func createProcess(process TextProcessContext) fs.FS {
	return fstest.MapFS{
		fmt.Sprintf("proc/%s/stat", process.Pid): {
			Data: []byte(statContent),
		},
		fmt.Sprintf("proc/%s/status", process.Pid): {
			Data: []byte(fmt.Sprintf(status, process.Name)),
		},
		fmt.Sprintf("proc/%s/cmdline", process.Pid): {
			Data: []byte(fmt.Sprintf(cmdline, process.Name, process.ConfigFileFlagKey, process.ConfigFilePath)),
		},
	}
}
