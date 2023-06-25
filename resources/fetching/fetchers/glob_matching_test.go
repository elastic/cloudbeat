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
	"github.com/stretchr/testify/suite"
	"os"
	"path/filepath"
	"testing"
)

// Based on https://github.com/yargevad/filepathx/blob/master/filepathx.go

type GlobMatcherTestSuite struct {
	suite.Suite
}

func TestGlobMatcherTestSuite(t *testing.T) {
	s := new(GlobMatcherTestSuite)

	suite.Run(t, s)
}

func (s *GlobMatcherTestSuite) TestGlobMatchingNonExistingPattern() {
	directoryName := "test-outer-dir"
	fileName := "file.txt"
	dir := createDirectoriesWithFiles(&s.Suite, "", directoryName, []string{fileName})
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(dir)

	filePath := filepath.Join(dir, fileName)
	matchedFiles, err := Glob(filePath + "/***")

	s.NoError(err)
	s.Nil(matchedFiles)
}

func (s *GlobMatcherTestSuite) TestGlobMatchingPathDoesNotExist() {
	directoryName := "test-outer-dir"
	fileName := "file.txt"
	dir := createDirectoriesWithFiles(&s.Suite, "", directoryName, []string{fileName})
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(dir)

	filePath := filepath.Join(dir, fileName)
	matchedFiles, err := Glob(filePath + "/abc")

	s.NoError(err)
	s.Nil(matchedFiles)
}

func (s *GlobMatcherTestSuite) TestGlobMatchingSingleFile() {
	directoryName := "test-outer-dir"
	fileName := "file.txt"
	dir := createDirectoriesWithFiles(&s.Suite, "", directoryName, []string{fileName})
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(dir)

	filePath := filepath.Join(dir, fileName)
	matchedFiles, err := Glob(filePath)

	s.NoError(err, "Glob could not fetch results")
	s.Equal(1, len(matchedFiles))
	s.Equal(matchedFiles[0], filePath)
}

func (s *GlobMatcherTestSuite) TestGlobDirectoryOnly() {
	directoryName := "test-outer-dir"
	fileName := "file.txt"
	dir := createDirectoriesWithFiles(&s.Suite, "", directoryName, []string{fileName})
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			s.Fail(err.Error())
		}
	}(dir)

	matchedFiles, err := Glob(dir)

	s.NoError(err, "Glob could not fetch results")
	s.Equal(1, len(matchedFiles))
	s.Equal(matchedFiles[0], dir)
}

func (s *GlobMatcherTestSuite) TestGlobOuterDirectoryOnly() {
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

	matchedFiles, err := Glob(outerDir + "/*")

	s.NoError(err, "Glob could not fetch results")
	s.Equal(2, len(matchedFiles))
	s.Equal(matchedFiles[0], filepath.Join(outerDir, outerFiles[0]))
	s.Equal(matchedFiles[1], innerDir)
}

func (s *GlobMatcherTestSuite) TestGlobDirectoryRecursively() {
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

	matchedFiles, err := Glob(outerDir + "/**")

	s.NoError(err, "Glob could not fetch results")
	s.Equal(6, len(matchedFiles))

	//When using glob matching recursively the first outer folder is being sent without a '/'
	s.Equal(matchedFiles[0], outerDir+"/")
	s.Equal(matchedFiles[1], filepath.Join(outerDir, outerFiles[0]))
	s.Equal(matchedFiles[2], innerDir)
	s.Equal(matchedFiles[3], filepath.Join(innerDir, innerFiles[0]))
	s.Equal(matchedFiles[4], innerInnerDir)
	s.Equal(matchedFiles[5], filepath.Join(innerInnerDir, innerInnerFiles[0]))
}
