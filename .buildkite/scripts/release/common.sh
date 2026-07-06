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

# update_version_beat
# Updates defaultBeatVersion in version/version.go to NEXT_CLOUDBEAT_VERSION and stages the file.
update_version_beat() {
    sed -i'' -E "s/const defaultBeatVersion = .*/const defaultBeatVersion = \"${NEXT_CLOUDBEAT_VERSION}\"/g" version/version.go
    git add version/version.go
    if ! git diff --cached --quiet; then
        git commit -m "Bump to ${NEXT_CLOUDBEAT_VERSION}"
    fi
}

# update_arm_templates <version>
# Updates ElasticAgentVersion in both Azure ARM templates and regenerates dev variants.
update_arm_templates() {
    local version="$1"
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
