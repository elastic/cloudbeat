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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/djherbis/times"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/cloudbeat/internal/resources/fetching"
	"github.com/elastic/cloudbeat/internal/resources/fetching/cycle"
	"github.com/elastic/cloudbeat/internal/resources/utils/user"
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
	Name      string    `mapstructure:"file.name,omitempty"`
	Mode      string    `mapstructure:"file.mode,omitempty"`
	Gid       string    `mapstructure:"file.gid,omitempty"`
	Uid       string    `mapstructure:"file.uid,omitempty"`
	Owner     string    `mapstructure:"file.owner,omitempty"`
	Group     string    `mapstructure:"file.group,omitempty"`
	Path      string    `mapstructure:"file.path,omitempty"`
	Inode     string    `mapstructure:"file.inode,omitempty"`
	Extension string    `mapstructure:"file.extension,omitempty"`
	Size      int64     `mapstructure:"file.size"`
	Type      string    `mapstructure:"file.type,omitempty"`
	Directory string    `mapstructure:"file.directory,omitempty"`
	Accessed  time.Time `mapstructure:"file.accessed"`
	Mtime     time.Time `mapstructure:"file.mtime"`
	Ctime     time.Time `mapstructure:"file.ctime"`
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
	osUser     user.OSUser
	resourceCh chan fetching.ResourceInfo
	patterns   []string
}

func NewFsFetcher(log *logp.Logger, ch chan fetching.ResourceInfo, patterns []string) *FileSystemFetcher {
	return &FileSystemFetcher{
		log:        log,
		resourceCh: ch,
		osUser:     user.NewOSUserUtil(),
		patterns:   patterns,
	}
}

func (f *FileSystemFetcher) Fetch(_ context.Context, cycleMetadata cycle.Metadata) error {
	f.log.Debug("Starting FileSystemFetcher.Fetch")

	// Input files might contain glob pattern
	for _, filePattern := range f.patterns {
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

			f.resourceCh <- fetching.ResourceInfo{Resource: resource, CycleMetadata: cycleMetadata}
		}
	}

	return nil
}

func (f *FileSystemFetcher) fetchSystemResource(filePath string) (*FSResource, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s, error: %w", filePath, err)
	}
	resourceInfo, _ := f.fromFileInfo(info, filePath)

	return resourceInfo, nil
}

func (f *FileSystemFetcher) fromFileInfo(info os.FileInfo, path string) (*FSResource, error) {
	if info == nil {
		return nil, nil
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, errors.New("not of type syscall.Stat_t")
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

	return &FSResource{
		EvalResource:  data,
		ElasticCommon: f.createFileCommonData(stat, data, path),
	}, nil
}

func (f *FileSystemFetcher) Stop() {
}

func (r FSResource) GetData() any {
	return r.EvalResource
}

func (r FSResource) GetIds() []string {
	return nil
}

func (r FSResource) GetElasticCommonData() (map[string]any, error) {
	m := map[string]any{}

	m["file.name"] = r.ElasticCommon.Name
	m["file.mode"] = r.ElasticCommon.Mode
	m["file.gid"] = r.ElasticCommon.Gid
	m["file.uid"] = r.ElasticCommon.Uid
	m["file.owner"] = r.ElasticCommon.Owner
	m["file.group"] = r.ElasticCommon.Group
	m["file.path"] = r.ElasticCommon.Path
	m["file.inode"] = r.ElasticCommon.Inode
	m["file.extension"] = r.ElasticCommon.Extension
	m["file.directory"] = r.ElasticCommon.Directory
	m["file.size"] = r.ElasticCommon.Size
	m["file.type"] = r.ElasticCommon.Type

	if !r.ElasticCommon.Accessed.IsZero() {
		m["file.accessed"] = r.ElasticCommon.Accessed
	}
	if !r.ElasticCommon.Mtime.IsZero() {
		m["file.mtime"] = r.ElasticCommon.Mtime
	}
	if !r.ElasticCommon.Ctime.IsZero() {
		m["file.ctime"] = r.ElasticCommon.Ctime
	}

	return m, nil
}

func (r FSResource) GetMetadata() (fetching.ResourceMetadata, error) {
	return fetching.ResourceMetadata{
		ID:      r.EvalResource.Path,
		Type:    FSResourceType,
		SubType: r.EvalResource.SubType,
		Name:    r.EvalResource.Path, // The Path from the container and not from the host
	}, nil
}

func getFSSubType(fileInfo os.FileInfo) string {
	if fileInfo.IsDir() {
		return DirSubType
	}
	return FileSubType
}

func (f *FileSystemFetcher) createFileCommonData(stat *syscall.Stat_t, data EvalFSResource, path string) FileCommonData {
	cd := FileCommonData{
		Name:      data.Name,
		Mode:      data.Mode,
		Gid:       data.Gid,
		Uid:       data.Uid,
		Owner:     data.Owner,
		Group:     data.Group,
		Path:      data.Path,
		Inode:     data.Inode,
		Extension: filepath.Ext(path),
		Directory: filepath.Dir(path),
		Size:      stat.Size,
		Type:      data.SubType,
	}

	t, err := times.Stat(path)
	if err != nil {
		f.log.Errorf("failed to get file time data (file %s), error - %s", path, err.Error())
	} else {
		cd.Accessed = t.AccessTime()
		cd.Mtime = t.ModTime()
		if t.HasChangeTime() {
			cd.Ctime = t.ChangeTime()
		}
	}

	return cd
}
