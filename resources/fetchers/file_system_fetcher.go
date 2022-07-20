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
	"github.com/djherbis/times"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/cloudbeat/resources/utils/user"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	FSResourceType = "file"
	FileSubType    = "file"
	DirSubType     = "directory"
	UserFile       = "/hostfs/etc/passwd"
	GroupFile      = "/hostfs/etc/group"
)

type EvalFSResource struct {
	Name    string `json:"name"`
	Mode    string `json:"mode"`
	Gid     string `json:"gid"`
	Uid     string `json:"uid"`
	Owner   string `json:"owner"`
	Group   string `json:"group"`
	Path    string `json:"path"`
	Inode   string `json:"inode"`
	SubType string `json:"sub_type"`
}

// FileCommonData According to https://www.elastic.co/guide/en/ecs/current/ecs-file.html
type FileCommonData struct {
	Name      string    `json:"name,omitempty"`
	Mode      string    `json:"mode,omitempty"`
	Gid       string    `json:"gid,omitempty"`
	Uid       string    `json:"uid,omitempty"`
	Owner     string    `json:"owner,omitempty"`
	Group     string    `json:"group,omitempty"`
	Path      string    `json:"path,omitempty"`
	Inode     string    `json:"inode,omitempty"`
	Extension string    `json:"extension,omitempty"`
	Size      int64     `json:"size"`
	Type      string    `json:"type,omitempty"`
	Directory string    `json:"directory,omitempty"`
	Accessed  time.Time `json:"accessed"`
	Mtime     time.Time `json:"mtime"`
	Ctime     time.Time `json:"ctime"`
}

type FSResource struct {
	EvalResource  EvalFSResource
	ElasticCommon FileCommonData
}

// FileSystemFetcher implement the Fetcher interface
// The FileSystemFetcher meant to fetch file/directories from the file system and ship it
// to the Cloudbeat
type FileSystemFetcher struct {
	log        *logp.Logger
	cfg        FileFetcherConfig
	osUser     user.OSUser
	resourceCh chan fetching.ResourceInfo
}

type FileFetcherConfig struct {
	fetching.BaseFetcherConfig
	Patterns []string `config:"patterns"` // Files and directories paths for the fetcher to extract info from
}

func (f *FileSystemFetcher) Fetch(ctx context.Context, cMetadata fetching.CycleMetadata) error {
	f.log.Debug("Starting FileSystemFetcher.Fetch")

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

			f.resourceCh <- fetching.ResourceInfo{Resource: resource, CycleMetadata: cMetadata}
		}
	}

	return nil
}

func (f *FileSystemFetcher) fetchSystemResource(filePath string) (FSResource, error) {

	info, err := os.Stat(filePath)
	if err != nil {
		return FSResource{}, fmt.Errorf("failed to fetch %s, error: %w", filePath, err)
	}
	resourceInfo, _ := f.fromFileInfo(info, filePath)

	return resourceInfo, nil
}

func (f *FileSystemFetcher) fromFileInfo(info os.FileInfo, path string) (FSResource, error) {

	if info == nil {
		return FSResource{}, nil
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return FSResource{}, errors.New("Not a syscall.Stat_t")
	}

	mod := strconv.FormatUint(uint64(info.Mode().Perm()), 8)
	uid := strconv.FormatUint(uint64(stat.Uid), 10)
	gid := strconv.FormatUint(uint64(stat.Gid), 10)
	inode := strconv.FormatUint(stat.Ino, 10)

	username, err := f.osUser.GetUserNameFromID(uid, UserFile)
	if err != nil {
		logp.Error(fmt.Errorf("failed to find username for uid %s, error - %+v", uid, err))
	}

	groupName, err := f.osUser.GetGroupNameFromID(gid, GroupFile)
	if err != nil {
		logp.Error(fmt.Errorf("failed to find groupname for gid %s, error - %+v", gid, err))
	}

	data := EvalFSResource{
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

	return FSResource{
		EvalResource:  data,
		ElasticCommon: enrichFileCommonData(stat, data, path),
	}, nil
}

func (f *FileSystemFetcher) Stop() {
}

func (r FSResource) GetData() any {
	return r.EvalResource
}

func (r FSResource) GetElasticCommonData() any {
	return r.ElasticCommon
}

func (r FSResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:        r.EvalResource.Path,
		Type:      FSResourceType,
		SubType:   r.EvalResource.SubType,
		Name:      r.EvalResource.Path, // The Path from the container and not from the host
		ECSFormat: FSResourceType,
	}, nil
}

func getFSSubType(fileInfo os.FileInfo) string {
	if fileInfo.IsDir() {
		return DirSubType
	}
	return FileSubType
}

func enrichFileCommonData(stat *syscall.Stat_t, data EvalFSResource, path string) FileCommonData {
	cd := FileCommonData{}
	if err := enrichFromFileResource(&cd, data); err != nil {
		logp.Error(fmt.Errorf("failed to decode data, Error: %v", err))
	}

	if err := enrichFileTimes(&cd, path); err != nil {
		logp.Error(err)
	}

	cd.Extension = filepath.Ext(path)
	cd.Directory = filepath.Dir(path)
	cd.Size = stat.Size
	cd.Type = data.SubType

	return cd
}

func enrichFileTimes(cd *FileCommonData, filepath string) error {
	t, err := times.Stat(filepath)
	if err != nil {
		return fmt.Errorf("failed to get file time data, error - %+v", err)
	}

	cd.Accessed = t.AccessTime()
	cd.Mtime = t.ModTime()

	if t.HasChangeTime() {
		cd.Ctime = t.ChangeTime()
	}

	return nil
}

func enrichFromFileResource(cd *FileCommonData, data EvalFSResource) error {
	return mapstructure.Decode(data, cd)
}
