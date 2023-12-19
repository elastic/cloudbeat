#!/bin/bash
set -euxo pipefail

function update_deps() {
    # Update dependencies in go.mod and go.sum
    go get -u "$1" && go mod tidy || return 1
    # Check if anything changed
    git diff --exit-code &>/dev/null && return 1
    # Build and test
    go build && go test -failfast || return 1
    # Add changes to git
    git add .
    return 0
}

go list -m -f '{{if not (or .Indirect .Main)}}{{.Path}}{{end}}' all | # List all direct dependencies
    grep -v 'github.com/elastic/beats/v7' |                           # Updated separately
    sort --random-sort |                                              # Avoid always having the same update order
    while read -r line; do
        update_deps "$line" || git restore . # Reset state if update fails to build and pass tests
    done
