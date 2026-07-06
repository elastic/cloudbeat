#!/usr/bin/env bash
set -euo pipefail

# Runs after the minor version bump PR is merged into main.
# Performs two operations:
#   1. Branch-out: create and push the new X.Y release branch from main
#   2. Hermit PR: sync CLOUDBEAT_VERSION in bin/hermit.hcl to match version.go
#
# Required env vars:
#   BRANCH      — new minor branch name (e.g. 9.3)
#   NEW_VERSION — first version on that branch (e.g. 9.3.0)
#   REPO        — repository name (e.g. cloudbeat)
#   WORKFLOW    — must be "minor"
: "${BRANCH:?BRANCH is required}"
: "${NEW_VERSION:?NEW_VERSION is required}"
: "${REPO:?REPO is required}"
: "${WORKFLOW:?WORKFLOW is required}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

GH_REPO="elastic/${REPO}"
DRY_RUN="${DRY_RUN:-false}"
HERMIT_BRANCH="sync-cloudbeat-version-$(date +%s)"

git fetch origin main
git checkout main

echo "--- Post-minor-merge parameters"
echo "  REPO:        ${REPO}"
echo "  BRANCH:      ${BRANCH}"
echo "  NEW_VERSION: ${NEW_VERSION}"
echo "  DRY_RUN:     ${DRY_RUN}"

setup_git_identity

branch_out() {
    echo "--- Creating release branch ${BRANCH} from main"

    if git ls-remote --exit-code --heads origin "${BRANCH}" >/dev/null 2>&1; then
        echo "Branch ${BRANCH} already exists on origin — skipping."
        return
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "Dry run: would create and push branch ${BRANCH}"
        return
    fi

    git checkout -b "${BRANCH}"
    git push origin "${BRANCH}"
    git checkout main
}

hermit_pr() {
    echo "--- Syncing CLOUDBEAT_VERSION in hermit.hcl to ${NEW_VERSION}"

    local existing_pr
    existing_pr=$(gh pr list --repo "${GH_REPO}" \
        --search "Sync CLOUDBEAT_VERSION in hermit.hcl to ${NEW_VERSION}" \
        --state open --json number --jq '.[0].number' 2>/dev/null || echo "")
    if [[ -n "${existing_pr}" ]]; then
        echo "Hermit PR #${existing_pr} already open — skipping."
        return
    fi

    git checkout -b "${HERMIT_BRANCH}" origin/main

    sed -i'' -E "s/\"CLOUDBEAT_VERSION\": \".*\"/\"CLOUDBEAT_VERSION\": \"${NEW_VERSION}\"/" bin/hermit.hcl
    git add bin/hermit.hcl

    if git diff --cached --quiet; then
        echo "hermit.hcl already at ${NEW_VERSION} — skipping."
        git checkout main
        return
    fi

    git commit -m "Sync CLOUDBEAT_VERSION in hermit.hcl to ${NEW_VERSION}"

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "Dry run: would push ${HERMIT_BRANCH} and open hermit PR"
        return
    fi

    git push origin "${HERMIT_BRANCH}"
    gh pr create \
        --repo "${GH_REPO}" \
        --head "${HERMIT_BRANCH}" \
        --base main \
        --title "Sync CLOUDBEAT_VERSION in hermit.hcl to ${NEW_VERSION}" \
        --body "Automated update of CLOUDBEAT_VERSION in hermit.hcl to match version.go"
}

branch_out
hermit_pr
