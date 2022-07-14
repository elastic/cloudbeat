package layerfs

import (
	"io/fs"
)

// DirFile wraps a fs.File and allows ReadDir to read entries from layerfs
// instead of the source layer dir.
// (implements fs.File and fs.ReadDirFile).
type DirFile struct {
	fs.File
	fs fs.FS

	layerFs *LayerFs
	name    string
}

// GetFs returns the source layer.
func (f *DirFile) GetFs() fs.FS {
	return f.fs
}

// ReadDir reads entries from the layerfs.
func (f *DirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	if n >= 0 {
		return nil, newError("could not ReadDir because n >= 0 is not supported", f.name)
	}

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, newError("could not ReadDir because dirFile does not point to a directory", f.name)
	}

	return f.layerFs.ReadDir(f.name)
}

// FileInfo wraps a fs.FileInfo and keeps a reference to the source layer.
// (implements fs.FileInfo).
type FileInfo struct {
	fs.FileInfo

	fs fs.FS
}

// GetFs returns the source layer.
func (f *FileInfo) GetFs() fs.FS {
	return f.fs
}

// DirEntry wraps a fs.DirEntry and keeps a reference to the source layer.
// (implements fs.DirEntry).
type DirEntry struct {
	fs.DirEntry

	fs fs.FS
}

// GetFs returns the source layer.
func (d *DirEntry) GetFs() fs.FS {
	return d.fs
}

// Info returns an on-demand constructed FileInfo pointing to the source layer.
func (d *DirEntry) Info() (fs.FileInfo, error) {
	info, err := d.DirEntry.Info()
	if err != nil {
		return nil, err
	}

	return FileInfo{
		info,
		d.fs,
	}, nil
}
