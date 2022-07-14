package bundle

import (
	"context"
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
func HostBundle(name string, bundle Bundle, ctx context.Context) error {
	if _, err := os.Stat(BundlesFolder); os.IsNotExist(err) {
		err := os.Mkdir(BundlesFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	bundlePath := filepath.Join(BundlesFolder, name)

	bundleBin, err := Build(bundle, ctx)
	if err != nil {
		return err
	}

	err = os.WriteFile(bundlePath, bundleBin, 0644)
	if err != nil {
		return err
	}

	return nil
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
