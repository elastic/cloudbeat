package bundle

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
)

type Writer struct {
	gzipWriter *gzip.Writer
	tarWriter  *tar.Writer
	writer     io.Writer
}

func NewWriter(w io.Writer) *Writer {
	gzipWriter := gzip.NewWriter(w)
	tarWriter := tar.NewWriter(gzipWriter)
	return &Writer{
		gzipWriter: gzipWriter,
		tarWriter:  tarWriter,
		writer:     w,
	}
}

func (b *Writer) Close() {
	b.tarWriter.Close()
	b.gzipWriter.Close()
}

func (b *Writer) AddFile(filePath string, reader io.Reader, length int64) error {
	header := &tar.Header{
		Name: filePath,
		Size: length,
		Mode: 0600,
	}

	err := b.tarWriter.WriteHeader(header)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not write header for file '%s', got error '%s'", filePath, err.Error()))
	}

	x, err := io.Copy(b.tarWriter, reader)
	if err != nil || x == 0 {
		return errors.New(fmt.Sprintf("Could not copy the file '%s' data to the tarball, got error '%s'", filePath, err.Error()))
	}

	return nil
}

func (b *Writer) AddFileFromDisk(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not open file '%s', got error '%s'", filePath, err.Error()))
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return errors.New(fmt.Sprintf("Could not get stat for file '%s', got error '%s'", filePath, err.Error()))
	}

	return b.AddFile(filePath, file, stat.Size())
}
