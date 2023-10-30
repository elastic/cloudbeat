package server

import (
	"context"
	"net/http"

	"github.com/elastic/csp-security-policies/bundle"
)

func HostCISKubernetes(path string) (http.Handler, error) {
	server := bundle.NewServer()
	err := bundle.HostBundle(path, bundle.CISKubernetesBundle(), context.Background())
	if err != nil {
		return nil, err
	}

	return server, nil
}

func HostEKSKubernetes(path string) (http.Handler, error) {
	server := bundle.NewServer()
	err := bundle.HostBundle(path, bundle.CISEksBundle(), context.Background())
	if err != nil {
		return nil, err
	}

	return server, nil
}
