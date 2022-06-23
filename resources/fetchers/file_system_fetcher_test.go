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
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/testhelper"
	"github.com/stretchr/testify/suite"
	"github.com/elastic/cloudbeat/resources/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
)

type FSFetcherTestSuite struct {
	suite.Suite

	log        *logp.Logger
	resourceCh chan fetching.ResourceInfo
}

func TestFSFetcherTestSuite(t *testing.T) {
	s := new(FSFetcherTestSuite)
	s.log = logp.NewLogger("cloudbeat_fs_fetcher_test_suite")

	if err := logp.TestingSetup(); err != nil {
		t.Error(err)
	}

	suite.Run(s)
}

func (s *FSFetcherTestSuite) SetupTest() {
	s.resourceCh = make(chan fetching.ResourceInfo, 50)
}

func (s *FSFetcherTestSuite) TearDownTest() {
	close(s.resourceCh)
}

func (s *FSFetcherTestSuite) TestFileFetcherFetchASingleFile() {
	directoryName := "test-outer-dir"
	files := []string{"file.txt"}
	dir := createDirectoriesWithFiles(s.Suite, "", directoryName, files)
	defer os.RemoveAll(dir)

	filePaths := []string{filepath.Join(dir, files[0])}
	cfg := FileFetcherConfig{
		Patterns: filePaths,
	}

	osUserMock := &utils.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher := FileSystemFetcher{
		log:    log,
		cfg:    cfg,
		osUser: osUserMock,
	}

	results, err := fileFetcher.Fetch(context.TODO())
	factory := FileSystemFactory{}
	fileFetcher, err := factory.CreateFrom(s.log, cfg, s.resourceCh)
	s.NoError(err)

	var results []fetching.ResourceInfo
	err = fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results = testhelper.CollectResources(s.resourceCh)

	s.Nil(err, "Fetcher was not able to fetch files from FS")
	s.Equal(1, len(results))

	fsResource := results[0].Resource.(FileSystemResource)
	s.Equal(files[0], fsResource.FileName)
	fsResource := results[0].(FileSystemResource)
	s.Equal(files[0], fsResource.Name)
	s.Equal("600", fsResource.Mode)
	s.Equal("root", fsResource.Owner)
	s.Equal("root", fsResource.Group)

	rMetadata := fsResource.GetMetadata()
	s.NotNil(rMetadata.ID)
	s.Equal(filePaths[0], rMetadata.Name)
	s.Equal(FileSubType, rMetadata.SubType)
	s.Equal(FSResourceType, rMetadata.Type)
}

func (s *FSFetcherTestSuite) TestFileFetcherFetchTwoPatterns() {
	outerDirectoryName := "test-outer-dir"
	outerFiles := []string{"output.txt", "output1.txt"}
	outerDir := createDirectoriesWithFiles(s.Suite, "", outerDirectoryName, outerFiles)
	defer os.RemoveAll(outerDir)

	paths := []string{filepath.Join(outerDir, outerFiles[0]), filepath.Join(outerDir, outerFiles[1])}
	cfg := FileFetcherConfig{
		Patterns: paths,
	}

	osUserMock := &utils.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil).Once()
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("etcd", nil).Once()
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil).Once()
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("etcd", nil).Once()

	fileFetcher, err := factory.CreateFrom(s.log, cfg, s.resourceCh)
	s.NoError(err)
	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher := FileSystemFetcher{
		log:    log,
		cfg:    cfg,
		osUser: osUserMock,
	}
	results, err := fileFetcher.Fetch(context.TODO())

	err = fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.Nil(err, "Fetcher was not able to fetch files from FS")
	s.Equal(2, len(results))

	firstFSResource := results[0].Resource.(FileSystemResource)
	s.Equal(outerFiles[0], firstFSResource.FileName)

	firstFSResource := results[0].(FileSystemResource)
	s.Equal(outerFiles[0], firstFSResource.Name)
	s.Equal("600", firstFSResource.Mode)
	s.Equal("root", firstFSResource.Owner)
	s.Equal("root", firstFSResource.Group)

	rMetadata := firstFSResource.GetMetadata()
	s.NotNil(rMetadata.ID)
	s.Equal(paths[0], rMetadata.Name)
	s.Equal(FileSubType, rMetadata.SubType)
	s.Equal(FSResourceType, rMetadata.Type)

	secFSResource := results[1].(FileSystemResource)
	s.Equal(outerFiles[1], secFSResource.Name)
	s.Equal("600", secFSResource.Mode)
	s.Equal("etcd", secFSResource.Owner)
	s.Equal("etcd", secFSResource.Group)
	secFSResource := results[1].Resource.(FileSystemResource)
	s.Equal(outerFiles[1], secFSResource.FileName)

	SecResMetadata := secFSResource.GetMetadata()
	s.NotNil(SecResMetadata.ID)
	s.Equal(paths[1], SecResMetadata.Name)
	s.Equal(FileSubType, SecResMetadata.SubType)
	s.Equal(FSResourceType, SecResMetadata.Type)
}

func (s *FSFetcherTestSuite) TestFileFetcherFetchDirectoryOnly() {
	directoryName := "test-outer-dir"
	files := []string{"file.txt"}
	dir := createDirectoriesWithFiles(s.Suite, "", directoryName, files)
	defer os.RemoveAll(dir)

	filePaths := []string{filepath.Join(dir)}
	cfg := FileFetcherConfig{
		Patterns: filePaths,
	}

	osUserMock := &utils.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("", errors.New("err"))
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("", errors.New("err"))

	fileFetcher, err := factory.CreateFrom(s.log, cfg, s.resourceCh)
	s.NoError(err)
	err = fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results := testhelper.CollectResources(s.resourceCh)
	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher := FileSystemFetcher{
		log:    log,
		cfg:    cfg,
		osUser: osUserMock,
	}
	results, err := fileFetcher.Fetch(context.TODO())

	s.Nil(err, "Fetcher was not able to fetch files from FS")
	s.Equal(1, len(results))

	fsResource := results[0].Resource.(FileSystemResource)
	expectedResult := filepath.Base(dir)
	rMetadata := fsResource.GetMetadata()

	s.Equal(expectedResult, fsResource.FileName)
	s.NotNil(rMetadata.ID)
	s.NotNil(rMetadata.Name)
	s.Equal(DirSubType, rMetadata.SubType)
	s.Equal(FSResourceType, rMetadata.Type)
	s.Equal(expectedResult, fsResource.Name)
	s.Equal("", fsResource.Owner)
	s.Equal("", fsResource.Group)
	assert.NotNil(rMetadata.ID)
	assert.NotNil(rMetadata.Name)
	s.Equal(DirSubType, rMetadata.SubType)
	s.Equal(FSResourceType, rMetadata.Type)
}

func (s *FSFetcherTestSuite) TestFileFetcherFetchOuterDirectoryOnly() {
	outerDirectoryName := "test-outer-dir"
	outerFiles := []string{"output.txt"}
	outerDir := createDirectoriesWithFiles(s.Suite, "", outerDirectoryName, outerFiles)
	defer os.RemoveAll(outerDir)

	innerDirectoryName := "test-inner-dir"
	innerFiles := []string{"innerFolderFile.txt"}
	innerDir := createDirectoriesWithFiles(s.Suite, outerDir, innerDirectoryName, innerFiles)

	path := []string{outerDir + "/*"}
	cfg := FileFetcherConfig{
		Patterns: path,
	}

	osUserMock := &utils.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher := FileSystemFetcher{
		log:    log,
		cfg:    cfg,
		osUser: osUserMock,
	}
	results, err := fileFetcher.Fetch(context.TODO())
	fileFetcher, err := factory.CreateFrom(s.log, cfg, s.resourceCh)
	s.NoError(err)
	err = fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.Nil(err, "Fetcher was not able to fetch files from FS")
	s.Equal(2, len(results))

	//All inner files should exist in the final result
	expectedResult := []string{"output.txt", filepath.Base(innerDir)}
	for i := 0; i < len(results); i++ {
		rMetadata := results[i].GetMetadata()
		fileSystemDataResources := results[i].Resource.(FileSystemResource)
		s.Contains(expectedResult, fileSystemDataResources.FileName)
		s.NotNil(rMetadata.SubType)
		s.NotNil(rMetadata.Name)
		s.NotNil(rMetadata.ID)
		s.Equal(FSResourceType, rMetadata.Type)
		s.NoError(err)
		fileSystemDataResources := results[i].(FileSystemResource)
		assert.Contains(expectedResult, fileSystemDataResources.Name)
		s.Equal("root", fileSystemDataResources.Owner)
		s.Equal("root", fileSystemDataResources.Group)
		assert.NotNil(rMetadata.SubType)
		assert.NotNil(rMetadata.Name)
		assert.NotNil(rMetadata.ID)
		s.Equal(FSResourceType, rMetadata.Type)
		assert.NoError(err)
	}
}

func (s *FSFetcherTestSuite) TestFileFetcherFetchDirectoryRecursively() {
	outerDirectoryName := "test-outer-dir"
	outerFiles := []string{"output.txt"}
	outerDir := createDirectoriesWithFiles(s.Suite, "", outerDirectoryName, outerFiles)
	defer os.RemoveAll(outerDir)

	innerDirectoryName := "test-inner-dir"
	innerFiles := []string{"innerFolderFile.txt"}
	innerDir := createDirectoriesWithFiles(s.Suite, outerDir, innerDirectoryName, innerFiles)

	innerInnerDirectoryName := "test-inner-inner-dir"
	innerInnerFiles := []string{"innerInnerFolderFile.txt"}
	innerInnerDir := createDirectoriesWithFiles(s.Suite, innerDir, innerInnerDirectoryName, innerInnerFiles)

	path := []string{outerDir + "/**"}
	cfg := FileFetcherConfig{
		Patterns: path,
	}
	osUserMock := &utils.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	fileFetcher, err := factory.CreateFrom(s.log, cfg, s.resourceCh)
	s.NoError(err)
	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher := FileSystemFetcher{
		log:    log,
		cfg:    cfg,
		osUser: osUserMock,
	}
	results, err := fileFetcher.Fetch(context.TODO())

	err = fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.Nil(err, "Fetcher was not able to fetch files from FS")
	s.Equal(6, len(results))

	directories := []string{filepath.Base(outerDir), filepath.Base(innerDir), filepath.Base(innerInnerDir)}
	allFilesName := append(append(append(innerFiles, directories...), outerFiles...), innerInnerFiles...)

	//All inner files should exist in the final result
	for i := 0; i < len(results); i++ {
		fileSystemDataResources := results[i].Resource.(FileSystemResource)
		rMetadata := results[i].GetMetadata()
		s.NotNil(rMetadata.SubType)
		s.NotNil(rMetadata.Name)
		s.NotNil(rMetadata.ID)
		s.Equal(FSResourceType, rMetadata.Type)
		s.NoError(err)
		s.Contains(allFilesName, fileSystemDataResources.FileName)
		assert.NotNil(rMetadata.SubType)
		assert.NotNil(rMetadata.Name)
		assert.NotNil(rMetadata.ID)
		s.Equal(FSResourceType, rMetadata.Type)
		assert.NoError(err)
		assert.Contains(allFilesName, fileSystemDataResources.Name)
		s.Equal("root", fileSystemDataResources.Owner)
		s.Equal("root", fileSystemDataResources.Group)
	}
}

// This function creates a new directory with files inside and returns the path of the new directory
func createDirectoriesWithFiles(s suite.Suite, dirPath string, dirName string, filesToWriteInDirectory []string) string {
	dirPath, err := ioutil.TempDir(dirPath, dirName)
	if err != nil {
		s.FailNow(err.Error())
	}
	for _, fileName := range filesToWriteInDirectory {
		file := filepath.Join(dirPath, fileName)
		s.Nil(ioutil.WriteFile(file, []byte("test txt\n"), 0600), "Could not able to write a new file")
	}
	return dirPath
}
