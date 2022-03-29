#!/usr/bin/env bash
set -euxo pipefail

source ./scripts/make/common.bash

jenkins_setup

cleanup() {
  rm -rf $TEMP_PYTHON_ENV
}
trap cleanup EXIT

make check-full
