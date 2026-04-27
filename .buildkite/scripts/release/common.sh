#!/usr/bin/env bash
# Sourced by bump scripts — do not execute directly.
# Expects callers to have validated: BRANCH, NEW_VERSION, REPO, WORKFLOW
# and to have set: BUMP_BRANCH, NEXT_CLOUDBEAT_VERSION, GH_REPO (=elastic/${REPO})

setup_git_identity() {
    git config --global user.email "cloudsecmachine@users.noreply.github.com"
    git config --global user.name "Cloud Security Machine"
}


