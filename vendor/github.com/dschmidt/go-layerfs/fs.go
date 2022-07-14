package layerfs

import (
	"io/fs"
)

// New creates a new LayerFs instance based on 0-n fs.FS layers.
func New(layers ...fs.FS) *LayerFs {
	return &LayerFs{layers: layers}
}

// LayerFs implements several interfaces from io/fs
// and delegates function calls sequentially to the underlying
// layers until one does not return an error. If all layers
// return errors, LayerFs returns an error
// In case of ReadDir LayerFs merges the DirEntry instances
// returned by the underlying layers.
type LayerFs struct {
	layers []fs.FS
}

// Open opens the named file (implements fs.FS).
func (fsys *LayerFs) Open(name string) (fs.File, error) {
	for _, layer := range fsys.layers {
		f, err := layer.Open(name)
		if err != nil {
			continue
		}

		return &DirFile{
			f,
			layer,
			fsys,
			name,
		}, nil
	}

	return nil, newError("could not Open", name)
}

// ReadFile reads the named file and returns its contents (implements fs.ReadFileFS).
func (fsys *LayerFs) ReadFile(name string) ([]byte, error) {
	for _, layer := range fsys.layers {
		file, err := fs.ReadFile(layer, name)
		if err != nil {
			continue
		}

		return file, nil
	}

	return nil, newError("could not ReadFile", name)
}

// ReadDir reads the named directory (implements fs.ReadDirFS).
func (fsys *LayerFs) ReadDir(name string) ([]fs.DirEntry, error) {
	entryMap := map[string]bool{}
	entries := make([]fs.DirEntry, 0)
	errorLayers := 0
	for _, layer := range fsys.layers {
		layerEntries, err := fs.ReadDir(layer, name)
		if err != nil {
			errorLayers++
			continue
		}
		for _, layerEntry := range layerEntries {
			_, ok := entryMap[layerEntry.Name()]
			if ok {
				continue
			}
			entryMap[layerEntry.Name()] = true
			lFsDirEntry := &DirEntry{
				layerEntry,
				layer,
			}
			entries = append(entries, lFsDirEntry)
		}
	}

	if errorLayers == len(fsys.layers) {
		return nil, newError("could not ReadDir", name)
	}

	return entries, nil
}

// Stat returns a FileInfo describing the file (implements fs.StatFS).
func (fsys *LayerFs) Stat(name string) (fs.FileInfo, error) {
	for _, layer := range fsys.layers {
		fi, err := fs.Stat(layer, name)
		if err != nil {
			continue
		}

		return FileInfo{
			fi,
			layer,
		}, nil
	}

	return nil, newError("could not Stat", name)
}
