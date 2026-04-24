#!/usr/bin/env bash
# Sourced by bump scripts — do not execute directly.
# Expects callers to have validated: BRANCH, NEW_VERSION, REPO, WORKFLOW
# and to have set: BUMP_BRANCH, NEXT_CLOUDBEAT_VERSION

setup_git_identity() {
    git config --global user.email "cloudsecmachine@users.noreply.github.com"
    git config --global user.name "Cloud Security Machine"
}

check_already_bumped() {
    if [[ "$(grep defaultBeatVersion version/version.go | cut -f2 -d '"')" == "${NEXT_CLOUDBEAT_VERSION}" ]]; then
        echo "INFO: version.go already at ${NEXT_CLOUDBEAT_VERSION} — nothing to do."
        exit 0
    fi
    local existing_pr
    existing_pr=$(gh pr list --repo "${REPO}" --head "${BUMP_BRANCH}" --state open --json number --jq '.[0].number' 2>/dev/null || echo "")
    if [[ -n "${existing_pr}" ]]; then
        echo "INFO: PR #${existing_pr} already open for ${BUMP_BRANCH} — skipping."
        exit 0
    fi
}

clear_stale_branch() {
    if git ls-remote --exit-code --heads origin "${BUMP_BRANCH}" &>/dev/null; then
        echo "Deleting stale remote branch: ${BUMP_BRANCH}"
        git push origin --delete "${BUMP_BRANCH}"
    fi
}
