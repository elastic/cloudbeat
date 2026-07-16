#!/usr/bin/env bash
set -euo pipefail

# Required env vars from PSI trigger params:
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
BUMP_BRANCH="bump-to-${NEW_VERSION}"
NEXT_CLOUDBEAT_VERSION="${NEW_VERSION}"
BACKPORT_LABEL="backport-v${BRANCH}.0"
DRY_RUN="${DRY_RUN:-false}"

git fetch origin main
git checkout main

CURRENT_CLOUDBEAT_VERSION=$(grep defaultBeatVersion version/version.go | cut -f2 -d '"')
POST_BUMP_MAIN_VERSION=$(next_minor_version "${NEW_VERSION}")

echo "--- Minor bump parameters"
echo "  REPO:                    ${REPO}"
echo "  BRANCH:                  ${BRANCH}"
echo "  CURRENT:                 ${CURRENT_CLOUDBEAT_VERSION}"
echo "  NEXT:                    ${NEXT_CLOUDBEAT_VERSION}"
echo "  POST_BUMP_MAIN_VERSION:  ${POST_BUMP_MAIN_VERSION}"
echo "  BACKPORT_LABEL:          ${BACKPORT_LABEL}"
echo "  DRY_RUN:                 ${DRY_RUN}"

# Idempotent no-op: if main has already been advanced to the post-bump minor,
# a previous run (or its post-minor-merge step) has completed this bump.
# Re-running would regress version.go/ARM templates and open a downgrade PR.
if [[ "${CURRENT_CLOUDBEAT_VERSION}" == "${POST_BUMP_MAIN_VERSION}" ]]; then
    echo "main is already at ${POST_BUMP_MAIN_VERSION} — minor bump for ${NEW_VERSION} has already completed. Idempotent no-op."
    exit 0
fi

setup_git_identity

update_mergify() {
    if grep -q "backport patches to ${BRANCH} branch" .mergify.yml; then
        echo "Mergify backport rule for ${BRANCH} already exists — skipping."
        return
    fi

    cat >>.mergify.yml <<EOF
  - name: backport patches to ${BRANCH} branch
    conditions:
      - merged
      - label=${BACKPORT_LABEL}
    actions:
      backport:
        assignees:
          - "{{ author }}"
        branches:
          - "${BRANCH}"
        labels:
          - "backport"
        title: "[{{ destination_branch }}](backport #{{ number }}) {{ title }}"
EOF

    git add .mergify.yml
    git commit -m "Add mergify backport rule for ${BRANCH}"
}

run_minor_bump() {
    pr_exists && return

    fail_if_stale_remote_branch "${BUMP_BRANCH}"

    git checkout -b "${BUMP_BRANCH}" origin/main

    update_version_beat
    update_arm_templates "${NEXT_CLOUDBEAT_VERSION}"
    update_mergify

    if no_new_commits origin/main; then
        echo "main is already at ${NEXT_CLOUDBEAT_VERSION} and mergify rule for ${BRANCH} already exists — nothing to bump."
        return
    fi

    local body
    body=$(render_template "${SCRIPT_DIR}/templates/pr-body-minor.md")

    # TODO: re-enable `--label "version-bump-auto-approve"` once the first
    # end-to-end manual run validates the auto-approve workflow + Mergify
    # auto-merge rule. Until then, the bump PR is created label-free so a
    # human reviews and merges it manually.
    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "--- Dry run: skipping push and PR creation"
        gh pr create \
            --repo "${GH_REPO}" \
            --head "${BUMP_BRANCH}" \
            --base main \
            --title "Bump cloudbeat version to ${NEXT_CLOUDBEAT_VERSION}" \
            --body "${body}" \
            --label "backport-skip" \
            --dry-run
        return
    fi

    git push --force origin "${BUMP_BRANCH}"
    gh pr create \
        --repo "${GH_REPO}" \
        --head "${BUMP_BRANCH}" \
        --base main \
        --title "Bump cloudbeat version to ${NEXT_CLOUDBEAT_VERSION}" \
        --body "${body}" \
        --label "backport-skip"
}

run_minor_bump
