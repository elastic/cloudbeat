#!/usr/bin/env bash
# Sourced by bump scripts — do not execute directly.
# Expects callers to have validated: BRANCH, NEW_VERSION, REPO, WORKFLOW
# and to have set: BUMP_BRANCH, NEXT_CLOUDBEAT_VERSION, GH_REPO (=elastic/${REPO})

# pr_exists
# Returns 0 (true) if an open PR already exists for BUMP_BRANCH, 1 otherwise.
pr_exists() {
    local existing_pr
    existing_pr=$(gh pr list --repo "${GH_REPO}" --head "${BUMP_BRANCH}" --state open \
        --json number --jq '.[0].number' 2>/dev/null || echo "")
    if [[ -n "${existing_pr}" ]]; then
        echo "INFO: PR #${existing_pr} already open for ${BUMP_BRANCH} — skipping."
        return 0
    fi
    return 1
}

setup_git_identity() {
    git config --global user.email "cloudsecmachine@users.noreply.github.com"
    git config --global user.name "Cloud Security Machine"
}

# is_downgrade <current> <target>
# Returns 0 (true) if <target> is strictly older than <current>, using semver
# ordering (via `sort -V`, which correctly handles multi-digit components).
# Returns 1 (false) if target == current or target > current.
is_downgrade() {
    local current="$1"
    local target="$2"
    if [[ "${current}" == "${target}" ]]; then
        return 1
    fi
    local highest
    highest=$(printf '%s\n%s\n' "${current}" "${target}" | sort -V | tail -1)
    [[ "${highest}" == "${current}" ]]
}

# update_version_beat
# Updates defaultBeatVersion in version/version.go to NEXT_CLOUDBEAT_VERSION and stages the file.
# Refuses to regress: if the current value is already >= NEXT_CLOUDBEAT_VERSION,
# leaves the file untouched. Prevents a stale retrigger from opening a downgrade PR.
update_version_beat() {
    local current
    current=$(grep defaultBeatVersion version/version.go | cut -f2 -d '"')
    if is_downgrade "${current}" "${NEXT_CLOUDBEAT_VERSION}"; then
        echo "Refusing to regress version/version.go: ${current} -> ${NEXT_CLOUDBEAT_VERSION} would be a downgrade."
        return
    fi
    sed -i'' -E "s/const defaultBeatVersion = .*/const defaultBeatVersion = \"${NEXT_CLOUDBEAT_VERSION}\"/g" version/version.go
    git add version/version.go
    if ! git diff --cached --quiet; then
        git commit -m "Bump to ${NEXT_CLOUDBEAT_VERSION}"
    fi
}

# update_arm_templates <version>
# Updates ElasticAgentVersion in both Azure ARM templates and regenerates dev variants.
# Refuses to regress: if the current default is already >= <version>, leaves the
# templates untouched.
update_arm_templates() {
    local version="$1"
    local current
    current=$(jq -r '.parameters.ElasticAgentVersion.defaultValue' deploy/azure/ARM-for-single-account.json)
    if is_downgrade "${current}" "${version}"; then
        echo "Refusing to regress Azure ARM templates: ${current} -> ${version} would be a downgrade."
        return
    fi
    echo "--- Update ARM templates to ${version}"
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"${version}\"" \
        deploy/azure/ARM-for-single-account.json >tmp.json && mv tmp.json deploy/azure/ARM-for-single-account.json
    jq --indent 4 ".parameters.ElasticAgentVersion.defaultValue = \"${version}\"" \
        deploy/azure/ARM-for-organization-account.json >tmp.json && mv tmp.json deploy/azure/ARM-for-organization-account.json
    ./deploy/azure/generate_dev_template.py --template-type single-account
    ./deploy/azure/generate_dev_template.py --template-type organization-account
    git add \
        deploy/azure/ARM-for-single-account.json \
        deploy/azure/ARM-for-single-account.dev.json \
        deploy/azure/ARM-for-organization-account.json \
        deploy/azure/ARM-for-organization-account.dev.json
    if ! git diff --cached --quiet; then
        git commit -m "Update Azure ARM templates to ${version}"
    fi
}

# no_new_commits <base_ref>
# Returns 0 (true) if HEAD has no changes relative to <base_ref>, i.e. the
# preceding update_* calls found nothing to bump. Callers should skip
# push/PR creation in that case — a branch identical to its base has no
# commits for `gh pr create` to open a PR with.
no_new_commits() {
    git diff --quiet "$1" HEAD
}

# fail_if_stale_remote_branch <branch>
# Exits with an actionable error if <branch> exists on origin from a
# closed-but-unmerged PR (or no PR at all). Only call this after pr_exists
# has confirmed there's no *open* PR for <branch> — a leftover branch like
# that is treated by GitHub as already having required checks, and blocks a
# fresh push with the same branch name ("4 of N required status checks are
# expected"), even though nothing is actually using it anymore.
#
# A branch whose PR was *merged* is not stale — it's evidence this exact
# bump already shipped, and no_new_commits() further down will correctly
# detect there's nothing left to do. Only flag the unmerged case here.
#
# We don't delete it automatically — that needs a human to confirm it's
# safe first — so this just fails clearly instead of letting the push fail
# further down with a more confusing error.
fail_if_stale_remote_branch() {
    local branch="$1"

    if ! git ls-remote --exit-code --heads origin "${branch}" >/dev/null 2>&1; then
        return
    fi

    local merged_pr
    merged_pr=$(gh pr list --repo "${GH_REPO}" --head "${branch}" --state merged \
        --json number --jq '.[0].number' 2>/dev/null || echo "")
    if [[ -n "${merged_pr}" ]]; then
        return
    fi

    echo "ERROR: ${branch} already exists on origin from a previous, closed (unmerged) run."
    echo "GitHub treats it as a stale protected branch and will reject a fresh push to it."
    echo "Delete it manually, then re-run this pipeline, e.g.:"
    echo "  gh api -X DELETE repos/${GH_REPO}/git/refs/heads/${branch}"
    exit 1
}

# next_minor_version <version>
# Given "X.Y.Z", returns the next minor version "X.(Y+1).0".
next_minor_version() {
    local version="$1"
    local major minor
    major=$(echo "${version}" | cut -d. -f1)
    minor=$(echo "${version}" | cut -d. -f2)
    echo "${major}.$((minor + 1)).0"
}

# render_template <path>
# Expands ${VAR} references in a template file using the caller's environment.
render_template() {
    local content
    # shellcheck disable=SC2016
    # Single quotes are intentional: we need a literal backslash passed to sed.
    content=$(sed 's/`/\\`/g' "$1")
    eval "cat <<__EOF__
${content}
__EOF__
"
}
