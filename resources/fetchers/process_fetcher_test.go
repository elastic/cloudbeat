package fetchers

import (
	"context"
	"fmt"
	"github.com/elastic/cloudbeat/config"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/json"
	"os"
	"path"
	"path/filepath"
	"testing"
)

const (
	hostfsDirectory = "hostfs"
	procfsDirectory = "proc"
	statContent     = `1167 (containerd-shim) S 1 1167 198 0 -1 1077952768 223005 9831 39 0 665 1329 8 10 20 0 12 0 76222 730476544 2268 18446744073709551615 1 1 0 0 0 0 1006249984 0 2143420159 0 0 0 17 2 0 0 0 0 0 0 0 0 0 0 0 0 0`
)

type TextProcessContext struct {
	Pid               string
	Name              string
	MountedPath       string
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
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Creating pseudo fs from getProcFixtures failed at fixtures/proc with error: %s", err)
	}
	defer os.RemoveAll(dir)
	mountedPath := getMountedPath(dir)

	testProcess := TextProcessContext{
		Pid:               "3",
		Name:              "kubelet",
		MountedPath:       mountedPath,
		ConfigFileFlagKey: "config",
		ConfigFilePath:    "test/path",
	}
	createProcess(t, testProcess)

	requiredProcesses := make(map[string]config.ProcessInputConfiguration)
	requiredProcesses["kubelet"] = config.ProcessInputConfiguration{CommandArguments: []string{"config"}}

	config := ProcessFetcherConfig{
		BaseFetcherConfig: BaseFetcherConfig{},
		Directory:         mountedPath,
		RequiredProcesses: requiredProcesses,
	}
	processesFetcher := NewProcessesFetcher(config)

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(fetchedResource), 1)

	processResource := fetchedResource[0].(ProcessResource)
	assert.Equal(t, processResource.PID, testProcess.Pid)
	assert.Equal(t, processResource.Stat.Name, "kubelet")
	assert.Contains(t, processResource.Cmd, "/usr/bin/kubelet")
}

func TestFetchWhenProcessDoesNotExist(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Creating pseudo fs from getProcFixtures failed at fixtures/proc with error: %s", err)
	}
	defer os.RemoveAll(dir)
	mountedPath := getMountedPath(dir)

	testProcess := TextProcessContext{
		Pid:               "3",
		Name:              "kubelet",
		MountedPath:       mountedPath,
		ConfigFileFlagKey: "config",
		ConfigFilePath:    "test/path",
	}
	createProcess(t, testProcess)

	requiredProcesses := make(map[string]config.ProcessInputConfiguration)
	requiredProcesses["someProcess"] = config.ProcessInputConfiguration{CommandArguments: []string{}}

	config := ProcessFetcherConfig{
		BaseFetcherConfig: BaseFetcherConfig{},
		Directory:         mountedPath,
		RequiredProcesses: requiredProcesses,
	}
	processesFetcher := NewProcessesFetcher(config)

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(fetchedResource), 0)
}

func TestFetchWhenNoFlagRequired(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Creating pseudo fs from getProcFixtures failed at fixtures/proc with error: %s", err)
	}
	defer os.RemoveAll(dir)
	mountedPath := getMountedPath(dir)

	testProcess := TextProcessContext{
		Pid:               "3",
		Name:              "kubelet",
		MountedPath:       mountedPath,
		ConfigFileFlagKey: "config",
		ConfigFilePath:    "test/path",
	}
	createProcess(t, testProcess)

	requiredProcesses := make(map[string]config.ProcessInputConfiguration)
	requiredProcesses["kubelet"] = config.ProcessInputConfiguration{CommandArguments: []string{}}

	config := ProcessFetcherConfig{
		BaseFetcherConfig: BaseFetcherConfig{},
		Directory:         mountedPath,
		RequiredProcesses: requiredProcesses,
	}
	processesFetcher := NewProcessesFetcher(config)

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(fetchedResource), 1)

	processResource := fetchedResource[0].(ProcessResource)
	assert.Equal(t, processResource.PID, testProcess.Pid)
	assert.Equal(t, processResource.Stat.Name, "kubelet")
	assert.Contains(t, processResource.Cmd, "/usr/bin/kubelet")
}

func TestFetchWhenFlagExistsWithConfigFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Creating pseudo fs from getProcFixtures failed at fixtures/proc with error: %s", err)
	}
	defer os.RemoveAll(dir)

	// Creating a yaml file for the process config
	configFlagKey := "config"
	yamlConfigName := "kubeletConfig.yaml"
	yamlConfig := ProcessConfigTestStruct{
		A: "A",
		B: 2,
	}
	yamlData, err := yaml.Marshal(&yamlConfig)
	configFilePath := filepath.Join(dir, hostfsDirectory, yamlConfigName)

	mountedPath := getMountedPath(dir)

	testProcess := TextProcessContext{
		Pid:               "3",
		Name:              "kubelet",
		MountedPath:       mountedPath,
		ConfigFileFlagKey: configFlagKey,
		ConfigFilePath:    yamlConfigName,
	}
	createProcess(t, testProcess)
	err = ioutil.WriteFile(configFilePath, yamlData, 0600)
	assert.Nil(t, err)

	requiredProcesses := make(map[string]config.ProcessInputConfiguration)
	requiredProcesses["kubelet"] = config.ProcessInputConfiguration{CommandArguments: []string{"config"}}

	config := ProcessFetcherConfig{
		BaseFetcherConfig: BaseFetcherConfig{},
		Directory:         mountedPath,
		RequiredProcesses: requiredProcesses,
	}
	processesFetcher := NewProcessesFetcher(config)

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	assert.Nil(t, err)
	assert.Equal(t, len(fetchedResource), 1)

	processResource := fetchedResource[0].(ProcessResource)
	assert.Equal(t, processResource.PID, testProcess.Pid)
	assert.Equal(t, processResource.Stat.Name, "kubelet")
	assert.Contains(t, processResource.Cmd, "/usr/bin/kubelet")

	configResource := processResource.Config[configFlagKey]
	var result ProcessConfigTestStruct
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &result})
	assert.Nil(t, err, "Could not decode process config result from yaml")
	err = decoder.Decode(configResource)
	assert.Nil(t, err, "Could not decode process config result from yaml")

	assert.Equal(t, result.A, yamlConfig.A)
	assert.Equal(t, result.B, yamlConfig.B)
}

func TestFetchWhenFlagExistsWithConfigFileFinal(t *testing.T) {

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Creating pseudo fs from getProcFixtures failed at fixtures/proc with error: %s", err)
	}
	defer os.RemoveAll(dir)

	testCases := []struct {
		configFileName string
		marshal        func(in interface{}) (out []byte, err error)
		configType     string
	}{
		{"kubeletConfig.yaml", yaml.Marshal, "yaml"},
		{"kubeletConfig.json", json.Marshal, "json"},
	}

	for _, test := range testCases {
		configFlagKey := "config"
		// Creating a yaml file for the process config
		processConfig := ProcessConfigTestStruct{
			A: "A",
			B: 2,
		}
		yamlData, err := test.marshal(&processConfig)
		configFilePath := filepath.Join(dir, hostfsDirectory, test.configFileName)

		mountedPath := getMountedPath(dir)

		testProcess := TextProcessContext{
			Pid:               "3",
			Name:              "kubelet",
			MountedPath:       mountedPath,
			ConfigFileFlagKey: configFlagKey,
			ConfigFilePath:    test.configFileName,
		}

		createProcess(t, testProcess)
		err = ioutil.WriteFile(configFilePath, yamlData, 0600)
		if err != nil {
			return
		}

		requiredProcesses := make(map[string]config.ProcessInputConfiguration)
		requiredProcesses["kubelet"] = config.ProcessInputConfiguration{CommandArguments: []string{"config"}}

		to_remove, err := test.marshal(&requiredProcesses)
		err = ioutil.WriteFile(filepath.Join(dir, hostfsDirectory, "a.yaml"), to_remove, 0600)
		if err != nil {
			return
		}

		config := ProcessFetcherConfig{
			BaseFetcherConfig: BaseFetcherConfig{},
			Directory:         mountedPath,
			RequiredProcesses: requiredProcesses,
		}
		processesFetcher := NewProcessesFetcher(config)

		fetchedResource, err := processesFetcher.Fetch(context.TODO())
		assert.Nil(t, err)
		assert.Equal(t, len(fetchedResource), 1)

		processResource := fetchedResource[0].(ProcessResource)
		assert.Equal(t, processResource.PID, testProcess.Pid)
		assert.Equal(t, processResource.Stat.Name, "kubelet")
		assert.Contains(t, processResource.Cmd, "/usr/bin/kubelet")

		configResource := processResource.Config[configFlagKey]
		var result ProcessConfigTestStruct
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: &result})
		assert.Nil(t, err, "Could not decode process config result from %s type", test.configType)
		err = decoder.Decode(configResource)
		assert.Nil(t, err, "Could not decode process config result from file %s", configFilePath)

		assert.Equal(t, result.A, processConfig.A)
		assert.Equal(t, result.B, processConfig.B)
	}
}

func createProcess(t *testing.T, process TextProcessContext) interface{} {
	processPath := path.Join(process.MountedPath, procfsDirectory, process.Pid)

	err := os.MkdirAll(processPath, 0755)
	if err != nil {
		t.Fatalf("Creating pseudo fs for a new process failed with error: %s", err)
	}

	filesToWrite := make(map[string]string)
	filesToWrite["stat"] = statContent
	filesToWrite["status"] = fmt.Sprintf(status, process.Name)
	filesToWrite["cmdline"] = fmt.Sprintf(cmdline, process.Name, process.ConfigFileFlagKey, process.ConfigFilePath)

	// creating all the relevant files for procfs to work
	for fileName, content := range filesToWrite {
		file := filepath.Join(processPath, fileName)
		err := ioutil.WriteFile(file, []byte(content), 0600)
		assert.NotNil(t, err)
	}
	return processPath
}

func getMountedPath(tempDir string) string {
	return path.Join(tempDir, hostfsDirectory)
}
