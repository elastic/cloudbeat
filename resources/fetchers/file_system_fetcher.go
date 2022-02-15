package fetchers

import (
	"context"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/cloudbeat/resources"
)

// FileSystemFetcher implement the resources.Fetcher interface
// The FileSystemFetcher meant to fetch file/directories from the file system and ship it
// to the Cloudbeat
type FileSystemFetcher struct {
	cfg FileFetcherConfig
}

type FileFetcherConfig struct {
	resources.BaseFetcherConfig
	Patterns []string `config:"patterns"` // Files and directories paths for the fetcher to extract info from
}

const (
	FileSystemType = "file-system"
)

func NewFileFetcher(cfg FileFetcherConfig) resources.Fetcher {
	return &FileSystemFetcher{
		cfg: cfg,
	}
}

func (f *FileSystemFetcher) Fetch(ctx context.Context) ([]resources.FetcherResult, error) {
	results := make([]resources.FetcherResult, 0)

	// Input files might contain glob pattern
	for _, filePattern := range f.cfg.Patterns {
		matchedFiles, err := Glob(filePattern)
		if err != nil {
			logp.Err("Failed to find matched glob for %s, error - %+v", filePattern, err)
		}
		for _, file := range matchedFiles {
			resource := f.fetchSystemResource(file)
			results = append(results, resources.FetcherResult{
				Type:     FileSystemType,
				Resource: resource,
			})
		}
	}
	return results, nil
}

func (f *FileSystemFetcher) fetchSystemResource(filePath string) interface{} {
	info, err := os.Stat(filePath)
	if err != nil {
		logp.Err("Failed to fetch %s, error - %+v", filePath, err)
		return nil
	}
	file := FromFileInfo(info, filePath)

	return file
}

func FromFileInfo(info os.FileInfo, path string) resources.FileSystemResource {
	if info == nil {
		return resources.FileSystemResource{}
	}

	stat := info.Sys().(*syscall.Stat_t)
	uid := stat.Uid
	gid := stat.Gid
	u := strconv.FormatUint(uint64(uid), 10)
	g := strconv.FormatUint(uint64(gid), 10)
	usr, _ := user.LookupId(u)
	group, _ := user.LookupGroupId(g)
	mod := strconv.FormatUint(uint64(info.Mode().Perm()), 8)

	data := resources.FileSystemResource{
		FileName: info.Name(),
		FileMode: mod,
		Uid:      usr.Name,
		Gid:      group.Name,
		Path:     path,
	}

	return data
}

func (f *FileSystemFetcher) Stop() {
}
