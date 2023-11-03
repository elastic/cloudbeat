#!/bin/bash
set -euxo pipefail

bin/hermit upgrade \
    aws-iam-authenticator \
    awscli \
    elastic-package \
    gcloud \
    gh \
    golangci-lint \
    jq \
    just \
    mage \
    opa \
    pre-commit \
    rain \
    regal \
    shellcheck \
    shfmt \
    yq
git status # git diff might not have output because only binaries change
