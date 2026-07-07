#!/usr/bin/env bash
set -euo pipefail

# Required env vars from PSI trigger params:
#   BRANCH      — release branch (e.g. 9.3)
#   NEW_VERSION — target patch version (e.g. 9.3.4)
#   REPO        — repository name (e.g. cloudbeat)
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

run_patch_bump() {
    pr_exists && return

    fail_if_stale_remote_branch "${BUMP_BRANCH}"

    git checkout -b "${BUMP_BRANCH}" "origin/${BASE_BRANCH}"

    update_version_beat

    if no_new_commits "origin/${BASE_BRANCH}"; then
        echo "${BASE_BRANCH} is already at ${NEXT_CLOUDBEAT_VERSION} — nothing to bump."
        return
    fi

    local body
    body=$(render_template "${SCRIPT_DIR}/templates/pr-body-patch.md")

    # TODO: re-enable `--label "version-bump-auto-approve"` once the first
    # end-to-end manual run validates the auto-approve workflow + Mergify
    # auto-merge rule. Until then, the bump PR is created label-free so a
    # human reviews and merges it manually.
    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "--- Dry run: skipping push and PR creation"
        gh pr create \
            --repo "${GH_REPO}" \
            --head "${BUMP_BRANCH}" \
            --base "${BASE_BRANCH}" \
            --title "Bump cloudbeat version ${BASE_BRANCH} to ${NEXT_CLOUDBEAT_VERSION}" \
            --body "${body}" \
            --label "backport-skip" \
            --dry-run
        return
    fi

    git push origin "${BUMP_BRANCH}"
    gh pr create \
        --repo "${GH_REPO}" \
        --head "${BUMP_BRANCH}" \
        --base "${BASE_BRANCH}" \
        --title "Bump cloudbeat version ${BASE_BRANCH} to ${NEXT_CLOUDBEAT_VERSION}" \
        --body "${body}" \
        --label "backport-skip"
}

run_patch_bump
