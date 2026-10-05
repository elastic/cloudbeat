#!/usr/bin/env bash
set -euo pipefail

# Runs after the minor version bump PR is merged into main.
# Performs two operations:
#   1. Branch-out: create and push the new X.Y release branch from main
#   2. Main bump: advance main's version to the next minor (e.g. 9.5.0 -> 9.6.0)
#
# Does NOT sync bin/hermit.hcl — that's owned by scripts/sync_internal_cloudbeat_version.sh
# (see .github/workflows/sync-internal-cloudbeat-version.yml), which gates the sync on the
# target snapshot actually being published so we never pin hermit to a version whose agent
# artifacts don't exist yet.
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
NEXT_MAIN_VERSION=$(next_minor_version "${NEW_VERSION}")
BUMP_BRANCH="bump-to-${NEXT_MAIN_VERSION}"

git fetch origin main
git checkout main

echo "--- Post-minor-merge parameters"
echo "  REPO:              ${REPO}"
echo "  BRANCH:            ${BRANCH}"
echo "  NEW_VERSION:       ${NEW_VERSION}"
echo "  NEXT_MAIN_VERSION: ${NEXT_MAIN_VERSION}"
echo "  DRY_RUN:           ${DRY_RUN}"

setup_git_identity

branch_out() {
    echo "--- Creating release branch ${BRANCH} from main"

    if git ls-remote --exit-code --heads origin "${BRANCH}" >/dev/null 2>&1; then
        echo "Branch ${BRANCH} already exists on origin — skipping."
        return
    fi

    # Sanity check: main should still be at NEW_VERSION when we cut the release
    # branch. If it has already been advanced (e.g. main == NEXT_MAIN_VERSION
    # because a previous bump_main_to_next_minor merged but branch_out never
    # ran), cutting the branch from current main would silently point ${BRANCH}
    # at the wrong code. Refuse and require manual intervention rather than
    # produce a wrong-content release branch.
    local main_version
    main_version=$(grep defaultBeatVersion version/version.go | cut -f2 -d '"')
    if [[ "${main_version}" != "${NEW_VERSION}" ]]; then
        echo "ERROR: main is at ${main_version}, expected ${NEW_VERSION}."
        echo "Refusing to cut ${BRANCH} — the resulting branch would not represent ${NEW_VERSION}'s code."
        echo "This means main advanced without ${BRANCH} being cut."
        echo "To recover: create ${BRANCH} manually from the commit where main was at ${NEW_VERSION}, then re-run this pipeline."
        exit 1
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "Dry run: would create and push branch ${BRANCH}"
        return
    fi

    git checkout -b "${BRANCH}"
    git push origin "${BRANCH}"
    git checkout main
}

bump_main_to_next_minor() {
    echo "--- Bumping main to next minor version ${NEXT_MAIN_VERSION}"

    pr_exists && return

    fail_if_stale_remote_branch "${BUMP_BRANCH}"

    git checkout -b "${BUMP_BRANCH}" origin/main

    NEXT_CLOUDBEAT_VERSION="${NEXT_MAIN_VERSION}"
    echo "  NEXT_CLOUDBEAT_VERSION: ${NEXT_CLOUDBEAT_VERSION}"
    update_version_beat
    update_arm_templates "${NEXT_MAIN_VERSION}"

    if git diff --quiet origin/main HEAD; then
        echo "main is already at ${NEXT_MAIN_VERSION} — skipping."
        git checkout main
        return
    fi

    local body
    body=$(render_template "${SCRIPT_DIR}/templates/pr-body-bump-main.md")

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "--- Dry run: skipping push and PR creation"
        gh pr create \
            --repo "${GH_REPO}" \
            --head "${BUMP_BRANCH}" \
            --base main \
            --title "Bump cloudbeat version to ${NEXT_MAIN_VERSION}" \
            --body "${body}" \
            --label "backport-skip" \
            --dry-run
        git checkout main
        return
    fi

    git push origin "${BUMP_BRANCH}"
    gh pr create \
        --repo "${GH_REPO}" \
        --head "${BUMP_BRANCH}" \
        --base main \
        --title "Bump cloudbeat version to ${NEXT_MAIN_VERSION}" \
        --body "${body}" \
        --label "backport-skip"
    git checkout main
}

branch_out
bump_main_to_next_minor
