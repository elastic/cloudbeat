#!/usr/bin/env bash
set -euxo pipefail
IMAGE_SUFFIX=$1
kind load image-archive /tmp/cloudbeat-${IMAGE_SUFFIX}.tar --name=kind-mono & kind load image-archive /tmp/pytest-${IMAGE_SUFFIX}.tar --name=kind-mono
