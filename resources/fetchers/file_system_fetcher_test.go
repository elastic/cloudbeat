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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/stretchr/testify/assert"
)

func TestFileFetcherFetchASingleFile(t *testing.T) {
	directoryName := "test-outer-dir"
	files := []string{"file.txt"}
	dir := createDirectoriesWithFiles(t, "", directoryName, files)
	defer os.RemoveAll(dir)

	filePaths := []string{filepath.Join(dir, files[0])}
	cfg := FileFetcherConfig{
		Patterns: filePaths,
	}

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	factory := FileSystemFactory{}
	fileFetcher, err := factory.CreateFrom(log, cfg, nil)
	assert.NoError(t, err)
	_ := fileFetcher.Fetch(context.TODO(), nil)

	assert.Nil(t, err, "Fetcher was not able to fetch files from FS")
	assert.Equal(t, 1, len(results))

	fsResource := results[0].(FileSystemResource)
	assert.Equal(t, files[0], fsResource.FileName)
	assert.Equal(t, "600", fsResource.FileMode)

	rMetadata := fsResource.GetMetadata()
	assert.NotNil(t, rMetadata.ID)
	assert.Equal(t, filePaths[0], rMetadata.Name)
	assert.Equal(t, FileSubType, rMetadata.SubType)
	assert.Equal(t, FSResourceType, rMetadata.Type)
}

func TestFileFetcherFetchTwoPatterns(t *testing.T) {
	outerDirectoryName := "test-outer-dir"
	outerFiles := []string{"output.txt", "output1.txt"}
	outerDir := createDirectoriesWithFiles(t, "", outerDirectoryName, outerFiles)
	defer os.RemoveAll(outerDir)

	paths := []string{filepath.Join(outerDir, outerFiles[0]), filepath.Join(outerDir, outerFiles[1])}
	cfg := FileFetcherConfig{
		Patterns: paths,
	}
	factory := FileSystemFactory{}

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher, err := factory.CreateFrom(log, cfg, nil)
	assert.NoError(t, err)
	_ := fileFetcher.Fetch(context.TODO(), nil)

	assert.Nil(t, err, "Fetcher was not able to fetch files from FS")
	assert.Equal(t, 2, len(results))

	firstFSResource := results[0].(FileSystemResource)
	assert.Equal(t, outerFiles[0], firstFSResource.FileName)
	assert.Equal(t, "600", firstFSResource.FileMode)

	rMetadata := firstFSResource.GetMetadata()
	assert.NotNil(t, rMetadata.ID)
	assert.Equal(t, paths[0], rMetadata.Name)
	assert.Equal(t, FileSubType, rMetadata.SubType)
	assert.Equal(t, FSResourceType, rMetadata.Type)

	secFSResource := results[1].(FileSystemResource)
	assert.Equal(t, outerFiles[1], secFSResource.FileName)
	assert.Equal(t, "600", secFSResource.FileMode)

	SecResMetadata := secFSResource.GetMetadata()
	assert.NotNil(t, SecResMetadata.ID)
	assert.Equal(t, paths[1], SecResMetadata.Name)
	assert.Equal(t, FileSubType, SecResMetadata.SubType)
	assert.Equal(t, FSResourceType, SecResMetadata.Type)
}

func TestFileFetcherFetchDirectoryOnly(t *testing.T) {
	directoryName := "test-outer-dir"
	files := []string{"file.txt"}
	dir := createDirectoriesWithFiles(t, "", directoryName, files)
	defer os.RemoveAll(dir)

	filePaths := []string{filepath.Join(dir)}
	cfg := FileFetcherConfig{
		Patterns: filePaths,
	}
	factory := FileSystemFactory{}

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher, err := factory.CreateFrom(log, cfg, nil)
	assert.NoError(t, err)
	_ := fileFetcher.Fetch(context.TODO(), nil)

	assert.Nil(t, err, "Fetcher was not able to fetch files from FS")
	assert.Equal(t, 1, len(results))

	fsResource := results[0].(FileSystemResource)
	expectedResult := filepath.Base(dir)
	rMetadata := fsResource.GetMetadata()

	assert.Equal(t, expectedResult, fsResource.FileName)
	assert.NotNil(t, rMetadata.ID)
	assert.NotNil(t, rMetadata.Name)
	assert.Equal(t, DirSubType, rMetadata.SubType)
	assert.Equal(t, FSResourceType, rMetadata.Type)
}

func TestFileFetcherFetchOuterDirectoryOnly(t *testing.T) {
	outerDirectoryName := "test-outer-dir"
	outerFiles := []string{"output.txt"}
	outerDir := createDirectoriesWithFiles(t, "", outerDirectoryName, outerFiles)
	defer os.RemoveAll(outerDir)

	innerDirectoryName := "test-inner-dir"
	innerFiles := []string{"innerFolderFile.txt"}
	innerDir := createDirectoriesWithFiles(t, outerDir, innerDirectoryName, innerFiles)

	path := []string{outerDir + "/*"}
	cfg := FileFetcherConfig{
		Patterns: path,
	}
	factory := FileSystemFactory{}

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher, err := factory.CreateFrom(log, cfg, nil)
	assert.NoError(t, err)
	_ := fileFetcher.Fetch(context.TODO(), nil)

	assert.Nil(t, err, "Fetcher was not able to fetch files from FS")
	assert.Equal(t, 2, len(results))

	//All inner files should exist in the final result
	expectedResult := []string{"output.txt", filepath.Base(innerDir)}
	for i := 0; i < len(results); i++ {
		rMetadata := results[i].GetMetadata()
		fileSystemDataResources := results[i].(FileSystemResource)
		assert.Contains(t, expectedResult, fileSystemDataResources.FileName)
		assert.NotNil(t, rMetadata.SubType)
		assert.NotNil(t, rMetadata.Name)
		assert.NotNil(t, rMetadata.ID)
		assert.Equal(t, FSResourceType, rMetadata.Type)
		assert.NoError(t, err)
	}
}

func TestFileFetcherFetchDirectoryRecursively(t *testing.T) {
	outerDirectoryName := "test-outer-dir"
	outerFiles := []string{"output.txt"}
	outerDir := createDirectoriesWithFiles(t, "", outerDirectoryName, outerFiles)
	defer os.RemoveAll(outerDir)

	innerDirectoryName := "test-inner-dir"
	innerFiles := []string{"innerFolderFile.txt"}
	innerDir := createDirectoriesWithFiles(t, outerDir, innerDirectoryName, innerFiles)

	innerInnerDirectoryName := "test-inner-inner-dir"
	innerInnerFiles := []string{"innerInnerFolderFile.txt"}
	innerInnerDir := createDirectoriesWithFiles(t, innerDir, innerInnerDirectoryName, innerInnerFiles)

	path := []string{outerDir + "/**"}
	cfg := FileFetcherConfig{
		Patterns: path,
	}
	factory := FileSystemFactory{}

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher, err := factory.CreateFrom(log, cfg, nil)
	assert.NoError(t, err)
	_ := fileFetcher.Fetch(context.TODO(), nil)

	assert.Nil(t, err, "Fetcher was not able to fetch files from FS")
	assert.Equal(t, 6, len(results))

	directories := []string{filepath.Base(outerDir), filepath.Base(innerDir), filepath.Base(innerInnerDir)}
	allFilesName := append(append(append(innerFiles, directories...), outerFiles...), innerInnerFiles...)

	//All inner files should exist in the final result
	for i := 0; i < len(results); i++ {
		fileSystemDataResources := results[i].(FileSystemResource)
		rMetadata := results[i].GetMetadata()
		assert.NotNil(t, rMetadata.SubType)
		assert.NotNil(t, rMetadata.Name)
		assert.NotNil(t, rMetadata.ID)
		assert.Equal(t, FSResourceType, rMetadata.Type)
		assert.NoError(t, err)
		assert.Contains(t, allFilesName, fileSystemDataResources.FileName)
	}
}

// This function creates a new directory with files inside and returns the path of the new directory
func createDirectoriesWithFiles(t *testing.T, dirPath string, dirName string, filesToWriteInDirectory []string) string {
	dirPath, err := ioutil.TempDir(dirPath, dirName)
	if err != nil {
		t.Fatal(err)
	}
	for _, fileName := range filesToWriteInDirectory {
		file := filepath.Join(dirPath, fileName)
		assert.Nil(t, ioutil.WriteFile(file, []byte("test txt\n"), 0600), "Could not able to write a new file")
	}
	return dirPath
}
