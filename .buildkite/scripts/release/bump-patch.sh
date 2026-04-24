#!/usr/bin/env bash
set -euo pipefail

# Required env vars from PSI trigger params:
#   BRANCH      — release branch (e.g. 9.3)
#   NEW_VERSION — target patch version (e.g. 9.3.4)
#   REPO        — repository name (e.g. elastic/cloudbeat)
#   WORKFLOW    — must be "patch"
: "${BRANCH:?BRANCH is required}"
: "${NEW_VERSION:?NEW_VERSION is required}"
: "${REPO:?REPO is required}"
: "${WORKFLOW:?WORKFLOW is required}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=common.sh
source "${SCRIPT_DIR}/common.sh"

BASE_BRANCH="${BRANCH}"
BUMP_BRANCH="bump-to-${NEW_VERSION}"
GH_REPO="elastic/${REPO}"

git fetch origin "${BASE_BRANCH}"
git checkout "${BASE_BRANCH}"

NEXT_CLOUDBEAT_VERSION="${NEW_VERSION}"
CURRENT_CLOUDBEAT_VERSION=$(grep defaultBeatVersion version/version.go | cut -f2 -d '"')

DRY_RUN="${DRY_RUN:-false}"

echo "--- Patch bump parameters"
echo "  REPO:     ${REPO}"
echo "  BRANCH:   ${BASE_BRANCH}"
echo "  CURRENT:  ${CURRENT_CLOUDBEAT_VERSION}"
echo "  NEXT:     ${NEXT_CLOUDBEAT_VERSION}"
echo "  DRY_RUN:  ${DRY_RUN}"

setup_git_identity

update_version_beat() {
    sed -i'' -E "s/const defaultBeatVersion = .*/const defaultBeatVersion = \"${NEXT_CLOUDBEAT_VERSION}\"/g" version/version.go
    git add version/version.go
    if ! git diff --cached --quiet; then
        git commit -m "Update version.go"
    fi
}

run_patch_bump() {
    git checkout -b "${BUMP_BRANCH}" "origin/${BASE_BRANCH}"

    update_version_beat

    if git diff "origin/${BASE_BRANCH}..HEAD" --quiet; then
        echo "No changes after bump — nothing to push."
        exit 0
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "--- Dry run: skipping push and PR creation"
        gh pr create \
            --repo "${GH_REPO}" \
            --head "${BUMP_BRANCH}" \
            --base "${BASE_BRANCH}" \
            --title "Bump ${BASE_BRANCH} to ${NEXT_CLOUDBEAT_VERSION}" \
            --body "Bump cloudbeat version - \`${NEXT_CLOUDBEAT_VERSION}\`" \
            --label "backport-skip" \
            --label "version-bump-auto-approve" \
            --dry-run
        return
    fi

    git push origin "${BUMP_BRANCH}"
    gh pr create \
        --repo "${GH_REPO}" \
        --head "${BUMP_BRANCH}" \
        --base "${BASE_BRANCH}" \
        --title "Bump ${BASE_BRANCH} to ${NEXT_CLOUDBEAT_VERSION}" \
        --body "Bump cloudbeat version - \`${NEXT_CLOUDBEAT_VERSION}\`" \
        --label "backport-skip" \
        --label "version-bump-auto-approve"
}

check_already_bumped
clear_stale_branch
run_patch_bump
