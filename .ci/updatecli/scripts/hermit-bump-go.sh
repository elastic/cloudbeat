#!/bin/bash

set -euxo pipefail

bin/hermit install "go-$1"
git status # git diff might not have output because only binaries change
