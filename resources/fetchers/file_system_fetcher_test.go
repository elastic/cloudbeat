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
	"github.com/elastic/cloudbeat/resources/utils/user"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/elastic/elastic-agent-libs/logp"
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

	suite.Run(t, s)
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
	dir := createDirectoriesWithFiles(&s.Suite, "", directoryName, files)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(dir)

	filePaths := []string{filepath.Join(dir, files[0])}
	cfg := FileFetcherConfig{
		Patterns: filePaths,
	}

	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher := FileSystemFetcher{
		log:        log,
		cfg:        cfg,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}

	var results []fetching.ResourceInfo
	err := fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results = testhelper.CollectResources(s.resourceCh)

	s.NoError(err, "Fetcher was not able to fetch files from FS")
	s.Equal(1, len(results))

	fsResource := results[0].Resource
	evalResource := fsResource.GetData().(EvalFSResource)

	s.Equal(files[0], evalResource.Name)
	s.Equal("600", evalResource.Mode)
	s.Equal("root", evalResource.Owner)
	s.Equal("root", evalResource.Group)

	rMetadata := fsResource.GetMetadata()
	s.NotNil(rMetadata.ID)
	s.Equal(filePaths[0], rMetadata.Name)
	s.Equal(FileSubType, rMetadata.SubType)
	s.Equal(FSResourceType, rMetadata.Type)
	s.NotNil(fsResource.GetElasticCommonData())
}

func (s *FSFetcherTestSuite) TestFileFetcherFetchTwoPatterns() {
	outerDirectoryName := "test-outer-dir"
	outerFiles := []string{"output.txt", "output1.txt"}
	outerDir := createDirectoriesWithFiles(&s.Suite, "", outerDirectoryName, outerFiles)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(outerDir)

	paths := []string{filepath.Join(outerDir, outerFiles[0]), filepath.Join(outerDir, outerFiles[1])}
	cfg := FileFetcherConfig{
		Patterns: paths,
	}

	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil).Once()
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("etcd", nil).Once()
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil).Once()
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("etcd", nil).Once()

	fileFetcher := FileSystemFetcher{
		log:        s.log,
		cfg:        cfg,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}

	err := fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.NoError(err, "Fetcher was not able to fetch files from FS")
	s.Equal(2, len(results))

	firstFSResource := results[0].Resource
	firstEvalResource := firstFSResource.GetData().(EvalFSResource)
	s.Equal(outerFiles[0], firstEvalResource.Name)
	s.Equal("600", firstEvalResource.Mode)
	s.Equal("root", firstEvalResource.Owner)
	s.Equal("root", firstEvalResource.Group)

	rMetadata := firstFSResource.GetMetadata()
	s.NotNil(rMetadata.ID)
	s.Equal(paths[0], rMetadata.Name)
	s.Equal(FileSubType, rMetadata.SubType)
	s.Equal(FSResourceType, rMetadata.Type)

	secFSResource := results[1].Resource
	secEvalResource := secFSResource.GetData().(EvalFSResource)
	s.Equal(outerFiles[1], secEvalResource.Name)
	s.Equal("600", secEvalResource.Mode)
	s.Equal("etcd", secEvalResource.Owner)
	s.Equal("etcd", secEvalResource.Group)

	SecResMetadata := secFSResource.GetMetadata()
	s.NotNil(SecResMetadata.ID)
	s.Equal(paths[1], SecResMetadata.Name)
	s.Equal(FileSubType, SecResMetadata.SubType)
	s.Equal(FSResourceType, SecResMetadata.Type)
}

func (s *FSFetcherTestSuite) TestFileFetcherFetchDirectoryOnly() {
	directoryName := "test-outer-dir"
	files := []string{"file.txt"}
	dir := createDirectoriesWithFiles(&s.Suite, "", directoryName, files)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(dir)

	filePaths := []string{filepath.Join(dir)}
	cfg := FileFetcherConfig{
		Patterns: filePaths,
	}

	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("", errors.New("err"))
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("", errors.New("err"))

	fileFetcher := FileSystemFetcher{
		log:        s.log,
		cfg:        cfg,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}
	err := fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.NoError(err, "Fetcher was not able to fetch files from FS")
	s.Equal(1, len(results))

	fsResource := results[0].Resource
	evalResource := fsResource.GetData().(EvalFSResource)
	expectedResult := filepath.Base(dir)
	rMetadata := fsResource.GetMetadata()

	s.NotNil(rMetadata.ID)
	s.NotNil(rMetadata.Name)
	s.Equal(DirSubType, rMetadata.SubType)
	s.Equal(FSResourceType, rMetadata.Type)
	s.Equal(expectedResult, evalResource.Name)
	s.Equal("", evalResource.Owner)
	s.Equal("", evalResource.Group)
	s.NotNil(rMetadata.ID)
	s.NotNil(rMetadata.Name)
	s.NotNil(fsResource.GetElasticCommonData())
}

func (s *FSFetcherTestSuite) TestFileFetcherFetchOuterDirectoryOnly() {
	outerDirectoryName := "test-outer-dir"
	outerFiles := []string{"output.txt"}
	outerDir := createDirectoriesWithFiles(&s.Suite, "", outerDirectoryName, outerFiles)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(outerDir)

	innerDirectoryName := "test-inner-dir"
	innerFiles := []string{"innerFolderFile.txt"}
	innerDir := createDirectoriesWithFiles(&s.Suite, outerDir, innerDirectoryName, innerFiles)

	path := []string{outerDir + "/*"}
	cfg := FileFetcherConfig{
		Patterns: path,
	}

	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher := FileSystemFetcher{
		log:        log,
		cfg:        cfg,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}

	err := fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.NoError(err, "Fetcher was not able to fetch files from FS")
	s.Equal(2, len(results))

	//All inner files should exist in the final result
	expectedResult := []string{"output.txt", filepath.Base(innerDir)}
	for i := 0; i < len(results); i++ {
		fsResource := results[i].Resource
		rMetadata := fsResource.GetMetadata()
		evalResource := fsResource.GetData().(EvalFSResource)

		s.Contains(expectedResult, evalResource.Name)
		s.NotNil(rMetadata.SubType)
		s.NotNil(rMetadata.Name)
		s.NotNil(rMetadata.ID)
		s.Equal("root", evalResource.Group)
		s.Equal("root", evalResource.Owner)
		s.Equal(FSResourceType, rMetadata.Type)
		s.NoError(err)
		s.NotNil(fsResource.GetElasticCommonData())
	}
}

func (s *FSFetcherTestSuite) TestFileFetcherFetchDirectoryRecursively() {
	outerDirectoryName := "test-outer-dir"
	outerFiles := []string{"output.txt"}
	outerDir := createDirectoriesWithFiles(&s.Suite, "", outerDirectoryName, outerFiles)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(outerDir)

	innerDirectoryName := "test-inner-dir"
	innerFiles := []string{"innerFolderFile.txt"}
	innerDir := createDirectoriesWithFiles(&s.Suite, outerDir, innerDirectoryName, innerFiles)

	innerInnerDirectoryName := "test-inner-inner-dir"
	innerInnerFiles := []string{"innerInnerFolderFile.txt"}
	innerInnerDir := createDirectoriesWithFiles(&s.Suite, innerDir, innerInnerDirectoryName, innerInnerFiles)

	path := []string{outerDir + "/**"}
	cfg := FileFetcherConfig{
		Patterns: path,
	}
	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	fileFetcher := FileSystemFetcher{
		log:        s.log,
		cfg:        cfg,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}

	err := fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.NoError(err, "Fetcher was not able to fetch files from FS")
	s.Equal(6, len(results))

	directories := []string{filepath.Base(outerDir), filepath.Base(innerDir), filepath.Base(innerInnerDir)}
	allFilesName := append(append(append(innerFiles, directories...), outerFiles...), innerInnerFiles...)

	//All inner files should exist in the final result
	for i := 0; i < len(results); i++ {
		fsResource := results[i].Resource
		rMetadata := fsResource.GetMetadata()
		evalResource := fsResource.GetData().(EvalFSResource)

		s.NotNil(rMetadata.SubType)
		s.NotNil(rMetadata.Name)
		s.NotNil(rMetadata.ID)
		s.Contains(allFilesName, evalResource.Name)
		s.Equal(FSResourceType, rMetadata.Type)

		s.NoError(err)
		s.Equal("root", evalResource.Owner)
		s.Equal("root", evalResource.Group)
		s.NotNil(fsResource.GetElasticCommonData())
	}
}

func (s *FSFetcherTestSuite) TestElasticCommonData() {
	directoryName := "test-outer-dir"
	files := []string{"file.txt"}
	dir := createDirectoriesWithFiles(&s.Suite, "", directoryName, files)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(dir)

	filePaths := []string{filepath.Join(dir, files[0])}
	cfg := FileFetcherConfig{
		Patterns: filePaths,
	}

	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	log := logp.NewLogger("cloudbeat_file_system_fetcher_test")
	fileFetcher := FileSystemFetcher{
		log:        log,
		cfg:        cfg,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}

	var results []fetching.ResourceInfo
	err := fileFetcher.Fetch(context.TODO(), fetching.CycleMetadata{})
	results = testhelper.CollectResources(s.resourceCh)

	s.NoError(err, "Fetcher was not able to fetch files from FS")
	s.Equal(1, len(results))

	fsResource := results[0].Resource
	fileInfo := fsResource.GetData().(EvalFSResource)
	cd := fsResource.GetElasticCommonData().(FileCommonData)

	s.NotNil(cd)
	s.Equal(fileInfo.Name, cd.Name)
	s.Equal(fileInfo.Owner, cd.Owner)
	s.Equal(fileInfo.Mode, cd.Mode)
	s.Equal(fileInfo.Group, cd.Group)
	s.Equal(fileInfo.Gid, cd.Gid)
	s.Equal(fileInfo.Uid, cd.Uid)
	s.Equal(fileInfo.Path, cd.Path)
	s.Equal(fileInfo.Inode, cd.Inode)
	s.Equal(filepath.Ext(files[0]), cd.Extension)
	s.Contains(cd.Directory, directoryName)
}

// This function creates a new directory with files inside and returns the path of the new directory
func createDirectoriesWithFiles(s *suite.Suite, dirPath string, dirName string, filesToWriteInDirectory []string) string {
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
