#!/usr/bin/env bash
set -euxo pipefail

kind load image-archive /tmp/cloudbeat-${{ env.CONTAINER_SUFFIX }}.tar --name=kind-mono & kind load image-archive /tmp/pytest-${{ env.CONTAINER_SUFFIX }}.tar --name=kind-mono
