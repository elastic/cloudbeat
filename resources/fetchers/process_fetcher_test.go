package fetchers

import (
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/suite"
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

type ProcessFetcherTestSuite struct {
	suite.Suite
}

func TestProcessFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessFetcherTestSuite))
}

var status = `Name:   %s`
var cmdline = `/usr/bin/%s --kubeconfig=/etc/kubernetes/kubelet.conf --%s=%s`

func (t *ProcessFetcherTestSuite) TestFetchWhenFlagExistsButNoFile() {
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
			"kubelet": {ConfigFileArguments: []string{"fetcherConfig"}}},
	}
	processesFetcher := &ProcessesFetcher{cfg: fetcherConfig, Fs: sysfs}

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	t.Nil(err)
	t.Equal(1, len(fetchedResource))

	processResource := fetchedResource[0].(ProcessResource)
	t.Equal(testProcess.Pid, processResource.PID)
	t.Equal("kubelet", processResource.Stat.Name)
	t.Contains(processResource.Cmd, "/usr/bin/kubelet")
}

func (t *ProcessFetcherTestSuite) TestFetchWhenProcessDoesNotExist() {
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
			"someProcess": {ConfigFileArguments: []string{"fetcherConfig"}}},
	}
	processesFetcher := &ProcessesFetcher{cfg: fetcherConfig, Fs: fsys}

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	t.Nil(err)
	t.Equal(0, len(fetchedResource))
}

func (t *ProcessFetcherTestSuite) TestFetchWhenNoFlagRequired() {
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
			"kubelet": {ConfigFileArguments: []string{}}},
	}
	processesFetcher := &ProcessesFetcher{cfg: fetcherConfig, Fs: fsys}

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	t.Nil(err)
	t.Equal(1, len(fetchedResource))

	processResource := fetchedResource[0].(ProcessResource)
	t.Equal(testProcess.Pid, processResource.PID)
	t.Equal("kubelet", processResource.Stat.Name)
	t.Contains(processResource.Cmd, "/usr/bin/kubelet")
}

func (t *ProcessFetcherTestSuite) TestFetchWhenFlagExistsWithConfigFile() {

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
				"kubelet": {ConfigFileArguments: []string{"fetcherConfig"}}},
		}
		processesFetcher := &ProcessesFetcher{cfg: fetcherConfig, Fs: sysfs}

		fetchedResource, err := processesFetcher.Fetch(context.TODO())
		t.Nil(err)
		t.Equal(1, len(fetchedResource))

		processResource := fetchedResource[0].(ProcessResource)
		t.Equal(testProcess.Pid, processResource.PID)
		t.Equal("kubelet", processResource.Stat.Name)
		t.Contains(processResource.Cmd, "/usr/bin/kubelet")

		configResource := processResource.ExternalData[configFlagKey]
		var result ProcessConfigTestStruct
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &result})
		t.Nil(err, "Could not decode process fetcherConfig result from %s type", test.configType)
		err = decoder.Decode(configResource)
		t.Nil(err, "Could not decode process fetcherConfig result from file %s", test.configFileName)

		t.Equal(processConfig.A, result.A)
		t.Equal(processConfig.B, result.B)
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
