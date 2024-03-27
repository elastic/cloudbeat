#!/bin/bash
set -euo pipefail

source scripts/common.sh

git config --global user.email "cloudsecmachine@users.noreply.github.com"
git config --global user.name "Cloud Security Machine"

branch_name="sync-cis-rule-templates"
repo="repos/elastic/integrations"
pr_number=$(gh api $repo/pulls -q ".[] | select(.head.ref == \"$branch_name\" and .state == \"open\") | .number")
templates_path="packages/cloud_security_posture/kibana/csp_rule_template"
manifest_path="packages/cloud_security_posture/manifest.yml"
changelog_path="packages/cloud_security_posture/changelog.yml"

# get new or existing sync-cis-rule-templates branch
cd ../integrations
if git fetch origin main "$branch_name" &>/dev/null; then
    git checkout "$branch_name"
    git reset origin/main --hard # reset to main, avoids conflicts
else
    git checkout -b "$branch_name" origin/main
fi

# generate the rule templates
cd ../cloudbeat
poetry run -C security-policies python security-policies/dev/generate_rule_templates.py

# commit and push the changes
cd ../integrations
git add "$templates_path"
git commit -m "Sync CIS rule templates"
git push origin "$branch_name" -f

# create a PR if it doesn't exist and assign labels
if [[ -z "$pr_number" ]]; then
    pr=$(gh api \
        --method POST \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        /$repo/pulls \
        -f title='[Cloud Security] Sync CIS rule templates' \
        -f body='' \
        -f head="$branch_name" \
        -f base='main')

    pr_number=$(echo "$pr" | jq -r '.html_url' | awk -F'/' '{print $NF}')

    gh api \
        --method POST \
        -H "Accept: application/vnd.github+json" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "/$repo/issues/$pr_number/labels" \
        -f "labels[]=Team:Cloud Security" -f "labels[]=enhancement"
fi

pr_url=$(gh api $repo/pulls -q ".[] | select(.head.ref == \"$branch_name\" and .state == \"open\") | .html_url")
bump_integration_version "$changelog_path" "$manifest_path" "$pr_url" "Add CIS rule templates"
git add "$changelog_path" "$manifest_path"
git commit -m "Bump integration version"
git push origin "$branch_name"

# create PR body
rows="$(git diff --name-only origin/main -- "$templates_path" | while read -r file; do jq --arg a "$pr_url/files#diff-$(echo -n "$file" | openssl dgst -sha256 | awk '{print $2}')" -r '.attributes.metadata.benchmark | "\(.id): \(.rule_number): \($a)"' "$file"; done | awk '{split($0, a, ": "); b[a[1]] = (b[a[1]] == "" ? "" : b[a[1]] ", ") "["a[2]"]""("a[3]")"} END {for (i in b) printf("| %s | %s |\n", i, b[i])}')"
body=$(
    cat <<EOF
Added rule templates for CIS rules:
| benchmark.id | benchmark.rule_number |
|--------------|-----------------------|
$rows
EOF
)

# update PR body
gh api \
    --method PATCH \
    -H "Accept: application/vnd.github+json" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "/repos/elastic/integrations/pulls/$pr_number" \
    -f body="$body"
