#!/bin/bash
set -euxo pipefail

source bin/activate-hermit

packages=(
    aws-iam-authenticator
    awscli
    elastic-package
    gcloud
    gh
    golangci-lint
    jq
    just
    kind
    mage
    opa
    pre-commit
    rain
    regal
    shellcheck
    shfmt
    yq
)

# `hermit upgrade` will only upgrade minor versions for packages that use semantic versioning. Not all of our hermit
# packages use semantic versioning and even for those that do, we want to be at least aware of new major versions and
# update to those if they don't cause too much of a breakage.
# Uninstalling and re-installing will always install to the latest major version.
hermit uninstall "${packages[@]}"
hermit install "${packages[@]}"
hermit list

# Update pre-commit hooks
pre-commit autoupdate
SKIP=update-rules-status,update-rule-metadata pre-commit run --all || true # Run to generate diffs, fix failures in PR

git status # git diff might not have output when only binaries change
