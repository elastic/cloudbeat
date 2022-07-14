package layerfs

import (
	"io/fs"
)

// GetLayerForDirEntry returns the source layer for a DirEntry.
func GetLayerForDirEntry(d fs.DirEntry) (fs.FS, error) {
	if entry, ok := d.(*DirEntry); ok {
		return entry.GetFs(), nil
	}

	info, err := d.Info()
	if err != nil {
		return nil, err
	}

	// Use indirection over Info() because WalkDir creates a statDirEntry
	// wrapper and there's no way we can inject our own type there
	// In contrast to that we can always provide our own fileInfo
	// and use that even from WalkDir callback.
	fileInfo, ok := info.(FileInfo)
	if !ok {
		return nil, newError("Could not assert DirEntry type", d.Name())
	}

	return fileInfo.GetFs(), nil
}
