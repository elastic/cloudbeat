package fetchers

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/elastic/beats/v7/libbeat/logp"
)

type FileSystemResource struct {
	FileName string `json:"filename"`
	FileMode string `json:"mode"`
	Gid      string `json:"gid"`
	Uid      string `json:"uid"`
	Path     string `json:"path"`
	Inode    string `json:"inode"`
}

// FileSystemFetcher implement the Fetcher interface
// The FileSystemFetcher meant to fetch file/directories from the file system and ship it
// to the Cloudbeat
type FileSystemFetcher struct {
	cfg FileFetcherConfig
}

type FileFetcherConfig struct {
	BaseFetcherConfig
	Patterns []string `config:"patterns"` // Files and directories paths for the fetcher to extract info from
}

const (
	FileSystemType = "file-system"
)

func NewFileFetcher(cfg FileFetcherConfig) Fetcher {
	return &FileSystemFetcher{
		cfg: cfg,
	}
}

func (f *FileSystemFetcher) Fetch(ctx context.Context) ([]FetchedResource, error) {
	results := make([]FetchedResource, 0)

	// Input files might contain glob pattern
	for _, filePattern := range f.cfg.Patterns {
		matchedFiles, err := Glob(filePattern)
		if err != nil {
			logp.Err("Failed to find matched glob for %s, error - %+v", filePattern, err)
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

	uid := stat.Uid
	gid := stat.Gid
	u := strconv.FormatUint(uint64(uid), 10)
	g := strconv.FormatUint(uint64(gid), 10)
	usr, _ := user.LookupId(u)
	group, _ := user.LookupGroupId(g)
	mod := strconv.FormatUint(uint64(info.Mode().Perm()), 8)
	inode := strconv.FormatUint(uint64(stat.Ino), 10)

	data := FileSystemResource{
		FileName: info.Name(),
		FileMode: mod,
		Uid:      usr.Name,
		Gid:      group.Name,
		Path:     path,
		Inode:    inode,
	}

	return data, nil
}

func (f *FileSystemFetcher) Stop() {
}

func (r FileSystemResource) GetID() string {
	return r.Inode
}

func (r FileSystemResource) GetData() interface{} {
	return r
}
