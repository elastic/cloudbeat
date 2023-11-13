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
    kind \
    mage \
    opa \
    pre-commit \
    rain \
    regal \
    shellcheck \
    shfmt \
    yq

# Update pre-commit hooks
pre-commit autoupdate
pre-commit run --all || true # Run to generate diffs, fix failures in PR

git status # git diff might not have output when only binaries change
