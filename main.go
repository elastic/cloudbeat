package main

import (
	"log"
	"net/http"

	"github.com/elastic/csp-security-policies/server"
)

func main() {
	server, _ := server.HostCISKubernetes("bun.tar.gz")
	log.Fatal(http.ListenAndServe(":8000", server))
}
