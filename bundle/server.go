package bundle

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
)

const (
	BundlesFolder    = "tmpBundles"
	BundlePathPrefix = "/bundles/"
)

// HostBundle writes the given bundle to the disk in order to serve it later
// Consequent calls to HostBundle with the same name will override the file
func HostBundle(name string, files map[string]string) error {
	if _, err := os.Stat(BundlesFolder); os.IsNotExist(err) {
		err := os.Mkdir(BundlesFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	bundlePath := filepath.Join(BundlesFolder, name)
	file, err := os.Create(bundlePath)
	if err != nil {
		return err
	}

	writer := NewWriter(file)
	defer writer.Close()

	for k, v := range files {
		reader := bytes.NewReader([]byte(v))
		writer.AddFile(k, reader, int64(len(v)))
	}

	return nil
}

func HostBundleWithDataYaml(name string, files map[string]string, dataYaml string) error {
	files["data.yaml"] = dataYaml

	err := HostBundle(name, files)
	return err
}

type Server struct {
	mux *http.ServeMux
}

func NewServer() *Server {
	mux := http.NewServeMux()
	staticFileServer := http.FileServer(http.Dir(BundlesFolder))
	mux.Handle(BundlePathPrefix, http.StripPrefix(BundlePathPrefix, staticFileServer))

	return &Server{
		mux: mux,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
