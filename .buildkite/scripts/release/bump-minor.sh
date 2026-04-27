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

echo "--- Minor bump parameters"
echo "  REPO:            ${REPO}"
echo "  BRANCH:          ${BRANCH}"
echo "  CURRENT:         ${CURRENT_CLOUDBEAT_VERSION}"
echo "  NEXT:            ${NEXT_CLOUDBEAT_VERSION}"
echo "  BACKPORT_LABEL:  ${BACKPORT_LABEL}"
echo "  DRY_RUN:         ${DRY_RUN}"

setup_git_identity

update_mergify() {
    local tmp_rule tmp_file
    tmp_rule=$(mktemp)
    tmp_file=$(mktemp)

    cat > "${tmp_rule}" << EOF
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

    while IFS= read -r line; do
        if [[ "${line}" == "  - name: auto-merge version bump PRs"* ]]; then
            cat "${tmp_rule}"
        fi
        printf '%s\n' "${line}"
    done < .mergify.yml > "${tmp_file}"

    mv "${tmp_file}" .mergify.yml
    rm "${tmp_rule}"
    git add .mergify.yml
}

run_minor_bump() {
    # If an open PR already exists for this bump branch, skip — don't overwrite
    # in-flight work. Individual file changes are idempotent via git diff.
    local existing_pr
    existing_pr=$(gh pr list --repo "${GH_REPO}" --head "${BUMP_BRANCH}" --state open \
        --json number --jq '.[0].number' 2>/dev/null || echo "")
    if [[ -n "${existing_pr}" ]]; then
        echo "INFO: PR #${existing_pr} already open for ${BUMP_BRANCH} — skipping."
        return
    fi

    if git ls-remote --exit-code --heads origin "${BUMP_BRANCH}" &>/dev/null; then
        echo "Deleting stale remote branch: ${BUMP_BRANCH}"
        git push origin --delete "${BUMP_BRANCH}"
    fi

    git checkout -b "${BUMP_BRANCH}" origin/main

    sed -i'' -E "s/const defaultBeatVersion = .*/const defaultBeatVersion = \"${NEXT_CLOUDBEAT_VERSION}\"/g" version/version.go
    git add version/version.go

    for f in \
        deploy/azure/ARM-for-single-account.json \
        deploy/azure/ARM-for-single-account.dev.json \
        deploy/azure/ARM-for-organization-account.json \
        deploy/azure/ARM-for-organization-account.dev.json; do
        jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"${NEXT_CLOUDBEAT_VERSION}\"" "$f" > tmp.json && mv tmp.json "$f"
        git add "$f"
    done

    update_mergify

    if git diff --cached --quiet; then
        echo "No changes after bump — nothing to push."
        return
    fi

    git commit -m "Bump to ${NEXT_CLOUDBEAT_VERSION}"

    local body
    body="Bump cloudbeat version to \`${NEXT_CLOUDBEAT_VERSION}\`

## DOD
- [ ] Add new GitHub label \`${BACKPORT_LABEL}\` to the repository"

    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "--- Dry run: skipping push and PR creation"
        gh pr create \
            --repo "${GH_REPO}" \
            --head "${BUMP_BRANCH}" \
            --base main \
            --title "Bump to ${NEXT_CLOUDBEAT_VERSION}" \
            --body "${body}" \
            --label "backport-skip" \
            --label "version-bump-auto-approve" \
            --dry-run
        return
    fi

    git push origin "${BUMP_BRANCH}"
    gh pr create \
        --repo "${GH_REPO}" \
        --head "${BUMP_BRANCH}" \
        --base main \
        --title "Bump to ${NEXT_CLOUDBEAT_VERSION}" \
        --body "${body}" \
        --label "backport-skip" \
        --label "version-bump-auto-approve"
}

run_minor_bump
