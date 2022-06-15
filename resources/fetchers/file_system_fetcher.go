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
	"os"
	"strconv"
	"syscall"

	"github.com/pkg/errors"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils"
)

const (
	FSResourceType = "file"
	FileSubType    = "file"
	DirSubType     = "directory"
	UserFile       = "/hostfs/etc/passwd"
	GroupFile      = "/hostfs/etc/group"
)

type FileSystemResource struct {
	Name    string `json:"name"`
	Mode    string `json:"mode"`
	Gid     uint32 `json:"gid"`
	Uid     uint32 `json:"uid"`
	Owner   string `json:"owner"`
	Group   string `json:"group"`
	Path    string `json:"path"`
	Inode   string `json:"inode"`
	SubType string `json:"sub_type"`
}

// FileSystemFetcher implement the Fetcher interface
// The FileSystemFetcher meant to fetch file/directories from the file system and ship it
// to the Cloudbeat
type FileSystemFetcher struct {
	log    *logp.Logger
	cfg    FileFetcherConfig
	OSUser utils.OSUser
}

type FileFetcherConfig struct {
	fetching.BaseFetcherConfig
	Patterns []string `config:"patterns"` // Files and directories paths for the fetcher to extract info from
}

func (f *FileSystemFetcher) Fetch(ctx context.Context) ([]fetching.Resource, error) {
	f.log.Debug("Starting FileSystemFetcher.Fetch")

	results := make([]fetching.Resource, 0)

	// Input files might contain glob pattern
	for _, filePattern := range f.cfg.Patterns {
		matchedFiles, err := Glob(filePattern)
		if err != nil {
			f.log.Errorf("Failed to find matched glob for %s, error: %+v", filePattern, err)
		}
		for _, file := range matchedFiles {
			resource, err := f.fetchSystemResource(file)
			if err != nil {
				f.log.Errorf("Unable to fetch fileSystemResource for file %v", file)
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
		return FileSystemResource{}, fmt.Errorf("failed to fetch %s, error: %w", filePath, err)
	}
	resourceInfo, _ := f.fromFileInfo(info, filePath)

	return resourceInfo, nil
}

func (f *FileSystemFetcher) fromFileInfo(info os.FileInfo, path string) (FileSystemResource, error) {

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
	username, err := f.OSUser.GetUserNameFromID(uid, UserFile)
	if err != nil {
		logp.Error(fmt.Errorf("failed to find username for uid %d, error - %+v", uid, err))
	}

	groupName, err := f.OSUser.GetGroupNameFromID(gid, GroupFile)
	if err != nil {
		logp.Error(fmt.Errorf("failed to find groupname for gid %d, error - %+v", gid, err))
	}

	data := FileSystemResource{
		Name:    info.Name(),
		Mode:    mod,
		Gid:     gid,
		Uid:     uid,
		Owner:   username,
		Group:   groupName,
		Path:    path,
		Inode:   inode,
		SubType: getFSSubType(info),
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
