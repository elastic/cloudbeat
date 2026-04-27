#!/usr/bin/env bash
set -euo pipefail

# Runs after the minor bump PR is merged (wait-for-merge block step approved).
# Opens the hermit PR on main and branches out ${BRANCH} from main.
#
# Required env vars from PSI trigger params:
#   BRANCH      — new minor branch name (e.g. 9.4)
#   NEW_VERSION — first version on that branch (e.g. 9.4.0)
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
HERMIT_BRANCH="bump-hermit-to-${NEW_VERSION}"
NEXT_CLOUDBEAT_VERSION="${NEW_VERSION}"
DRY_RUN="${DRY_RUN:-false}"

echo "--- Minor finalize parameters"
echo "  REPO:        ${REPO}"
echo "  BRANCH:      ${BRANCH}"
echo "  NEW_VERSION: ${NEXT_CLOUDBEAT_VERSION}"
echo "  DRY_RUN:     ${DRY_RUN}"

setup_git_identity

open_hermit_pr() {
    if git ls-remote --exit-code --heads origin "${HERMIT_BRANCH}" &>/dev/null; then
        echo "Deleting stale remote branch: ${HERMIT_BRANCH}"
        git push origin --delete "${HERMIT_BRANCH}"
    fi

    git fetch origin main
    git checkout -b "${HERMIT_BRANCH}" origin/main

    sed -i'' -E 's/"CLOUDBEAT_VERSION": "[^"]*"/"CLOUDBEAT_VERSION": "'"${NEXT_CLOUDBEAT_VERSION}"'"/' bin/hermit.hcl
    git add bin/hermit.hcl

    if git diff --cached --quiet; then
        echo "hermit.hcl already at ${NEXT_CLOUDBEAT_VERSION} — skipping hermit PR."
        return
    fi

    git commit -m "Bump hermit CLOUDBEAT_VERSION to ${NEXT_CLOUDBEAT_VERSION}"

    local body="Update hermit CLOUDBEAT_VERSION to \`${NEXT_CLOUDBEAT_VERSION}\`.

> Merge only after the \`${NEXT_CLOUDBEAT_VERSION}\` snapshot build is available."

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "--- Dry run: skipping push and PR creation for hermit bump"
        gh pr create \
            --repo "${GH_REPO}" \
            --head "${HERMIT_BRANCH}" \
            --base main \
            --title "Bump hermit to ${NEXT_CLOUDBEAT_VERSION}" \
            --body "${body}" \
            --label "backport-skip" \
            --dry-run
        return
    fi

    git push origin "${HERMIT_BRANCH}"
    gh pr create \
        --repo "${GH_REPO}" \
        --head "${HERMIT_BRANCH}" \
        --base main \
        --title "Bump hermit to ${NEXT_CLOUDBEAT_VERSION}" \
        --body "${body}" \
        --label "backport-skip"
}

branch_out() {
    git fetch origin main

    if git ls-remote --exit-code --heads origin "${BRANCH}" &>/dev/null; then
        echo "INFO: Branch ${BRANCH} already exists on remote — skipping."
        return
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "--- Dry run: would create branch ${BRANCH} from origin/main"
        return
    fi

    git checkout -b "${BRANCH}" origin/main
    git push origin "${BRANCH}"
    echo "Branch ${BRANCH} created from origin/main."
}

open_hermit_pr
branch_out
