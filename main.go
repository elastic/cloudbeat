package main

import (
	"os"

	"github.com/elastic/cloudbeat/cmd"

	_ "github.com/elastic/cloudbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
