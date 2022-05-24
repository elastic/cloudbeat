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
	"fmt"
	"github.com/elastic/cloudbeat/resources/utils"
	"os"
	"strconv"
	"syscall"

	"github.com/pkg/errors"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
)

const (
	FSResourceType = "file"
	FileSubType    = "file"
	DirSubType     = "directory"
	UserFile       = "/hostfs/etc/passwd"
	GroupFile      = "/hostfs/etc/group"
)

type FileSystemResource struct {
	FileName string `json:"filename"`
	FileMode string `json:"mode"`
	Gid      string `json:"gid"`
	Uid      string `json:"uid"`
	Path     string `json:"path"`
	Inode    string `json:"inode"`
	SubType  string `json:"sub_type"`
}

// FileSystemFetcher implement the Fetcher interface
// The FileSystemFetcher meant to fetch file/directories from the file system and ship it
// to the Cloudbeat
type FileSystemFetcher struct {
	cfg FileFetcherConfig
}

type FileFetcherConfig struct {
	fetching.BaseFetcherConfig
	Patterns []string `config:"patterns"` // Files and directories paths for the fetcher to extract info from
}

func (f *FileSystemFetcher) Fetch(ctx context.Context) ([]fetching.Resource, error) {
	logp.L().Debug("file fetcher starts to fetch data")
	results := make([]fetching.Resource, 0)

	// Input files might contain glob pattern
	for _, filePattern := range f.cfg.Patterns {
		matchedFiles, err := Glob(filePattern)
		if err != nil {
			logp.Error(fmt.Errorf("failed to find matched glob for %s, error - %+v", filePattern, err))
		}
		for _, file := range matchedFiles {
			resource, err := f.fetchSystemResource(file)
			if err != nil {
				logp.Err("Unable to fetch fileSystemResource for file: %v", file)
				continue
			}
			results = append(results, resource)
		}
	}
	return results, nil
}

func (f *FileSystemFetcher) fetchSystemResource(filePath string) (FileSystemResource, error) {

	info, err := os.Stat(filePath)
	if err != nil {
		err := fmt.Errorf("failed to fetch %s, error - %+v", filePath, err)
		return FileSystemResource{}, err
	}
	resourceInfo, _ := FromFileInfo(info, filePath)

	return resourceInfo, nil
}

func FromFileInfo(info os.FileInfo, path string) (FileSystemResource, error) {

	if info == nil {
		return FileSystemResource{}, nil
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return FileSystemResource{}, errors.New("Not a syscall.Stat_t")
	}

	mod := strconv.FormatUint(uint64(info.Mode().Perm()), 8)
	inode := strconv.FormatUint(stat.Ino, 10)

	uid := stat.Uid
	gid := stat.Gid
	username, err := utils.GetUserNameFromID(uid, UserFile)
	if err != nil {
		logp.Error(fmt.Errorf("failed to find username for uid %d, error - %+v", uid, err))
	}

	groupName, err := utils.GetGroupNameFromID(gid, GroupFile)
	if err != nil {
		logp.Error(fmt.Errorf("failed to find groupname for gid %d, error - %+v", gid, err))
	}

	data := FileSystemResource{
		FileName: info.Name(),
		FileMode: mod,
		Uid:      username,
		Gid:      groupName,
		Path:     path,
		Inode:    inode,
		SubType:  getFSSubType(info),
	}

	return data, nil
}

func (f *FileSystemFetcher) Stop() {
}

func (r FileSystemResource) GetData() interface{} {
	return r
}

func (r FileSystemResource) GetMetadata() fetching.ResourceMetadata {
	return fetching.ResourceMetadata{
		ID:      r.Path,
		Type:    FSResourceType,
		SubType: r.SubType,
		Name:    r.Path, // The Path from the container and not from the host
	}
}

func getFSSubType(fileInfo os.FileInfo) string {
	if fileInfo.IsDir() {
		return DirSubType
	}
	return FileSubType
}
