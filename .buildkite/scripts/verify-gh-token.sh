#!/usr/bin/env bash
set -euo pipefail

# ---------------------------------------------------------------------------
# Verification script for task-01: confirm the wolfi agent image has gh CLI
# and that the ephemeral vault-github-token can authenticate and has the
# permissions needed to create PRs in elastic/cloudbeat.
#
# Run this step BEFORE wiring the full bump script into the pipeline.
# ---------------------------------------------------------------------------

REPO="elastic/cloudbeat"

echo "--- Check gh CLI is available"
if ! command -v gh &>/dev/null; then
  echo "ERROR: gh CLI not found in PATH"
  echo ""
  echo "Image info:"
  cat /etc/os-release 2>/dev/null || uname -a
  echo ""
  echo "PATH: $PATH"
  echo ""
  echo "ACTION REQUIRED: install gh CLI in this image or switch to an image that includes it."
  exit 1
fi
gh --version

echo "--- Verify authenticated actor"
gh auth status

echo "--- Configure git identity"
git config --global user.email "cloudsecmachine@users.noreply.github.com"
git config --global user.name "Cloud Security Machine"

TEST_BRANCH="verify-gh-token/${BUILDKITE_BUILD_NUMBER:-local}"

echo "--- Create test branch ${TEST_BRANCH}"
git checkout -b "${TEST_BRANCH}"

echo "--- gh pr create"
touch .gh-pr-create-test-file
git add .gh-pr-create-test-file
git commit -m "chore: test gh pr create permissions"
git push origin "${TEST_BRANCH}"
gh pr create --repo "${REPO}" \
  --head "${TEST_BRANCH}" \
  --title "chore: verify ephemeral token — delete me" \
  --body "Automated verification that the ephemeral vault token can create PRs. Safe to close and delete." \
  --base main

echo ""
echo "All checks passed:"
echo "  gh CLI       : $(gh --version | head -1)"
echo "  pr create DR : ok"
