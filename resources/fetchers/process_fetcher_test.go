package fetchers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

const (
	hostfsDirectory = "hostfs"
	procfsDirectory = "proc"
)

func TestMyFirstTest(t *testing.T) {
	pid := "3"
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Creating pseudo fs from getProcFixtures failed at fixtures/proc with error: %s", err)
	}
	defer os.RemoveAll(dir)
	mountedPath := getMountedPath(dir)
	processPath := createProcess(t, mountedPath, pid,nil)
	//fs := getProcFixtures(t, dir)
	//_, err = fs.Proc(pid)
	//if err != nil {
	//	t.Fatal(err)
	//}

	config := ProcessFetcherConfig{
		BaseFetcherConfig: BaseFetcherConfig{},
		Directory:         mountedPath,
		RequiredProcesses: nil,
	}
	processesFetcher := NewProcessesFetcher(config)

	fetchedResource, err := processesFetcher.Fetch(context.TODO())
	assert.NotNil(t, fetchedResource)
	assert.NotNil(t, processPath)
}

// This function creates a new directory with files inside and returns the path of the new directory
func createProcess(t *testing.T, mountPath string, processId string, filesToWriteInDirectory []string) string {
	processPath := path.Join(mountPath,hostfsDirectory, procfsDirectory, processId)
	os.Mkdir(processPath, 0755)

	for _, fileName := range filesToWriteInDirectory {
		file := filepath.Join(processPath, fileName)
		assert.Nil(t, ioutil.WriteFile(file, []byte("test txt\n"), 0600), "Could not able to write a new file")
	}
	return processPath
}

func getMountedPath(tempDir string) string  {
	return path.Join(tempDir, hostfsDirectory)
}
