#!/usr/bin/env bash
set -euo pipefail

# Usage: approve-pr.sh <pr-number>
# Approves a bot PR as the GitHub Actions bot when it needs no human review.
# Called by .github/workflows/auto-approve-prs.yml. Approves:
#   - Renovate PRs labeled renovate-auto-approve. renovate.json decides which
#     updates get the label, and Renovate merges them itself.
#   - Patch version bump PRs labeled version-bump-auto-approve, opened by
#     .buildkite/scripts/release/bump-patch.sh on a release branch, that only
#     change defaultBeatVersion in version/version.go.
# Anything else is left for a human. If the PR has one of these labels but does
# not qualify, the script comments on the PR with the reason.
#
# Env: GH_REPO (e.g. elastic/cloudbeat), GH_TOKEN, DRY_RUN=true to only print what it would do.

PR="${1:?usage: approve-pr.sh <pr-number>}"
: "${GH_REPO:?GH_REPO is required}"
DRY_RUN="${DRY_RUN:-false}"

pr=$(gh pr view "${PR}" --json author,labels,baseRefName,files,latestReviews,comments)

has_label() {
    jq -e --arg label "$1" 'any(.labels[]; .name == $label)' <<<"${pr}" >/dev/null
}

# Runs a command that writes to the PR, or only prints it when DRY_RUN=true.
run() {
    if [[ "${DRY_RUN}" == "true" ]]; then
        echo "Dry run: $*"
    else
        "$@"
    fi
}

# The check_* functions print why the PR does not qualify, or nothing if it does.
check_renovate() {
    has_label renovate-auto-approve || echo "missing the renovate-auto-approve label"
}

# Qualifies if the PR targets a release branch and its only change is the
# defaultBeatVersion line in version/version.go.
check_patch_bump() {
    local base files bumped
    has_label version-bump-auto-approve || {
        echo "missing the version-bump-auto-approve label"
        return
    }
    base=$(jq -r .baseRefName <<<"${pr}")
    [[ "${base}" =~ ^[0-9]+\.[0-9]+$ ]] || {
        echo "base branch ${base} is not a release branch"
        return
    }
    files=$(jq -r '[.files[] | "\(.path) +\(.additions) -\(.deletions)"] | join(", ")' <<<"${pr}")
    [[ "${files}" == "version/version.go +1 -1" ]] || {
        echo "changes more than one line in version/version.go, changed: ${files}"
        return
    }
    bumped=$(gh pr diff "${PR}" | grep -cE '^\+const defaultBeatVersion = "[0-9]+\.[0-9]+\.[0-9]+"$' || true)
    [[ "${bumped}" == "1" ]] || echo "the version/version.go change is not a defaultBeatVersion bump"
}

author=$(jq -r .author.login <<<"${pr}")
case "${author}" in
app/elastic-renovate-prod)
    reason="Renovate update labeled renovate-auto-approve"
    problem=$(check_renovate)
    ;;
app/elastic-vault-github-plugin-prod)
    reason="patch version bump that only changes defaultBeatVersion in version/version.go"
    problem=$(check_patch_bump)
    ;;
*)
    problem="author ${author} does not qualify for auto-approval"
    ;;
esac

if [[ -n "${problem}" ]]; then
    message="Not auto-approved, this PR needs a review: ${problem}."
    echo "PR #${PR}: ${message}"
    # Comment only on PRs that asked for auto-approval, and only once per reason:
    # the workflow runs again on every label and push.
    if { has_label renovate-auto-approve || has_label version-bump-auto-approve; } &&
        ! jq -e --arg body "${message}" 'any(.comments[]; .body == $body)' <<<"${pr}" >/dev/null; then
        run gh pr comment "${PR}" --body "${message}"
    fi
    exit 0
fi

if jq -e 'any(.latestReviews[]; .author.login == "github-actions" and .state == "APPROVED")' <<<"${pr}" >/dev/null; then
    echo "PR #${PR} is already approved by github-actions."
    exit 0
fi

run gh pr review "${PR}" --approve --body "Automated approval: ${reason}."
