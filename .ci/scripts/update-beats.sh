#!/bin/bash
set -euo pipefail

BEATS_VERSION=${1:?Missing version argument}
go get "github.com/elastic/beats/v7@$BEATS_VERSION"
go mod tidy
