#!/usr/bin/env bash
set -euxo pipefail

rm -rf /tmp/.buildx-cache
mv /tmp/.buildx-cache-new /tmp/.buildx-cache
