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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/utils/testhelper"
	"github.com/elastic/cloudbeat/internal/resources/utils/user"
)

type FSFetcherTestSuite struct {
	suite.Suite

	resourceCh chan fetching.ResourceInfo
}

func TestFSFetcherTestSuite(t *testing.T) {
	s := new(FSFetcherTestSuite)

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

	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	fileFetcher := FileSystemFetcher{
		log:        testhelper.NewLogger(s.T()),
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
		patterns:   filePaths,
	}

	var results []fetching.ResourceInfo
	err := fileFetcher.Fetch(context.TODO(), cycle.Metadata{})
	results = testhelper.CollectResources(s.resourceCh)

	s.Require().NoError(err, "Fetcher was not able to fetch files from FS")
	s.Len(results, 1)

	fsResource := results[0].Resource
	evalResource := fsResource.GetData().(EvalFSResource)

	s.Equal(files[0], evalResource.Name)
	s.Equal("600", evalResource.Mode)
	s.Equal("root", evalResource.Owner)
	s.Equal("root", evalResource.Group)

	rMetadata, err := fsResource.GetMetadata()
	s.Require().NoError(err)
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

	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil).Once()
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("etcd", nil).Once()
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil).Once()
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("etcd", nil).Once()

	fileFetcher := FileSystemFetcher{
		log:        testhelper.NewLogger(s.T()),
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
		patterns:   paths,
	}

	err := fileFetcher.Fetch(context.TODO(), cycle.Metadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.Require().NoError(err, "Fetcher was not able to fetch files from FS")
	s.Len(results, 2)

	firstFSResource := results[0].Resource
	firstEvalResource := firstFSResource.GetData().(EvalFSResource)
	s.Equal(outerFiles[0], firstEvalResource.Name)
	s.Equal("600", firstEvalResource.Mode)
	s.Equal("root", firstEvalResource.Owner)
	s.Equal("root", firstEvalResource.Group)

	rMetadata, err := firstFSResource.GetMetadata()
	s.Require().NoError(err)
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

	secResMetadata, err := secFSResource.GetMetadata()
	s.Require().NoError(err)
	s.NotNil(secResMetadata.ID)
	s.Equal(paths[1], secResMetadata.Name)
	s.Equal(FileSubType, secResMetadata.SubType)
	s.Equal(FSResourceType, secResMetadata.Type)
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

	filePaths := []string{filepath.Clean(dir)}
	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("", errors.New("err"))
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("", errors.New("err"))

	fileFetcher := FileSystemFetcher{
		log:        testhelper.NewLogger(s.T()),
		patterns:   filePaths,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}
	err := fileFetcher.Fetch(context.TODO(), cycle.Metadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.Require().NoError(err, "Fetcher was not able to fetch files from FS")
	s.Len(results, 1)

	fsResource := results[0].Resource
	evalResource := fsResource.GetData().(EvalFSResource)
	expectedResult := filepath.Base(dir)
	rMetadata, err := fsResource.GetMetadata()

	s.Require().NoError(err)
	s.NotNil(rMetadata.ID)
	s.NotNil(rMetadata.Name)
	s.Equal(DirSubType, rMetadata.SubType)
	s.Equal(FSResourceType, rMetadata.Type)
	s.Equal(expectedResult, evalResource.Name)
	s.Empty(evalResource.Owner)
	s.Empty(evalResource.Group)
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
	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	fileFetcher := FileSystemFetcher{
		log:        testhelper.NewLogger(s.T()),
		patterns:   path,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}

	err := fileFetcher.Fetch(context.TODO(), cycle.Metadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.Require().NoError(err, "Fetcher was not able to fetch files from FS")
	s.Len(results, 2)

	// All inner files should exist in the final result
	expectedResult := []string{"output.txt", filepath.Base(innerDir)}
	for i := range results {
		fsResource := results[i].Resource
		rMetadata, err := fsResource.GetMetadata()
		s.Require().NoError(err)
		evalResource := fsResource.GetData().(EvalFSResource)

		s.Contains(expectedResult, evalResource.Name)
		s.NotNil(rMetadata.SubType)
		s.NotNil(rMetadata.Name)
		s.NotNil(rMetadata.ID)
		s.Equal("root", evalResource.Group)
		s.Equal("root", evalResource.Owner)
		s.Equal(FSResourceType, rMetadata.Type)
		s.Require().NoError(err)
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
	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	fileFetcher := FileSystemFetcher{
		log:        testhelper.NewLogger(s.T()),
		patterns:   path,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}

	err := fileFetcher.Fetch(context.TODO(), cycle.Metadata{})
	results := testhelper.CollectResources(s.resourceCh)

	s.Require().NoError(err, "Fetcher was not able to fetch files from FS")
	s.Len(results, 6)

	directories := []string{filepath.Base(outerDir), filepath.Base(innerDir), filepath.Base(innerInnerDir)}
	allFilesName := append(append(append(innerFiles, directories...), outerFiles...), innerInnerFiles...)

	// All inner files should exist in the final result
	for i := range results {
		fsResource := results[i].Resource
		rMetadata, err := fsResource.GetMetadata()
		evalResource := fsResource.GetData().(EvalFSResource)

		s.Require().NoError(err)
		s.NotNil(rMetadata.SubType)
		s.NotNil(rMetadata.Name)
		s.NotNil(rMetadata.ID)
		s.Contains(allFilesName, evalResource.Name)
		s.Equal(FSResourceType, rMetadata.Type)

		s.Require().NoError(err)
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
	osUserMock := &user.MockOSUser{}
	osUserMock.EXPECT().GetUserNameFromID(mock.Anything, mock.Anything).Return("root", nil)
	osUserMock.EXPECT().GetGroupNameFromID(mock.Anything, mock.Anything).Return("root", nil)

	fileFetcher := FileSystemFetcher{
		log:        testhelper.NewLogger(s.T()),
		patterns:   filePaths,
		osUser:     osUserMock,
		resourceCh: s.resourceCh,
	}

	var results []fetching.ResourceInfo
	err := fileFetcher.Fetch(context.TODO(), cycle.Metadata{})
	results = testhelper.CollectResources(s.resourceCh)

	s.Require().NoError(err, "Fetcher was not able to fetch files from FS")
	s.Len(results, 1)

	fsResource := results[0].Resource
	fileInfo := fsResource.GetData().(EvalFSResource)
	fileCd, err := fsResource.GetElasticCommonData()
	s.Require().NoError(err)

	s.NotNil(fileCd)
	s.Equal(fileInfo.Name, fileCd["file.name"])
	s.Equal(fileInfo.Owner, fileCd["file.owner"])
	s.Equal(fileInfo.Mode, fileCd["file.mode"])
	s.Equal(fileInfo.Group, fileCd["file.group"])
	s.Equal(fileInfo.Gid, fileCd["file.gid"])
	s.Equal(fileInfo.Uid, fileCd["file.uid"])
	s.Equal(fileInfo.Path, fileCd["file.path"])
	s.Equal(fileInfo.Inode, fileCd["file.inode"])
	s.Equal(filepath.Ext(files[0]), fileCd["file.extension"])
	s.Contains(fileCd["file.directory"], directoryName)
}

// This function creates a new directory with files inside and returns the path of the new directory
func createDirectoriesWithFiles(s *suite.Suite, dirPath string, dirName string, filesToWriteInDirectory []string) string {
	dirPath, err := os.MkdirTemp(dirPath, dirName)
	if err != nil {
		s.FailNow(err.Error())
	}
	for _, fileName := range filesToWriteInDirectory {
		file := filepath.Join(dirPath, fileName)
		s.Require().NoError(os.WriteFile(file, []byte("test txt\n"), 0600), "Could not able to write a new file")
	}
	return dirPath
}
