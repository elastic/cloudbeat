package bundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTarGzWriter(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		files map[string]string
	}{
		{
			map[string]string{},
		},
		{
			map[string]string{
				"file.txt": "some text",
			},
		},
		{
			map[string]string{
				"folder/file.rego":  "package authz",
				"folder/route.rego": "package route",
				"a.txt":             "some text",
			},
		},
	}

	for _, test := range tests {
		buf := &bytes.Buffer{}
		writer := NewWriter(buf)

		for k, v := range test.files {
			reader := bytes.NewReader([]byte(v))
			writer.AddFile(k, reader, int64(len(v)))
		}
		writer.Close()

		result, err := extractGzipFiles(buf)
		assert.NoError(err)
		assert.Equal(test.files, result)
	}
}

func extractGzipFiles(gz io.Reader) (map[string]string, error) {
	gzipFiles := map[string]string{}
	gzf, err := gzip.NewReader(gz)
	if err != nil {
		return nil, err
	}

	tarReader := tar.NewReader(gzf)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			buf := &bytes.Buffer{}
			_, err = io.Copy(buf, tarReader)
			gzipFiles[header.Name] = buf.String()

		default:
			return nil, errors.New("could not read type")
		}
	}

	return gzipFiles, nil
}
