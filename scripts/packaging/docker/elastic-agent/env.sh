#!/usr/bin/env bash

set -eu

VERSION=$(grep defaultBeatVersion version/version.go | cut -f2 -d "\"")
DEFAULT_IMAGE_TAG="${DEFAULT_IMAGE_TAG:-${VERSION}-SNAPSHOT}"
BASE_IMAGE="${BASE_IMAGE:-docker.elastic.co/beats/elastic-agent:$DEFAULT_IMAGE_TAG}"

echo "BASE_IMAGE=${BASE_IMAGE}"
