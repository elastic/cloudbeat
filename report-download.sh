#!/usr/bin/env bash

# mkdir -p tests/allure/results
mkdir -p artifacts && cd artifacts

# artifacts_url=${{ github.event.workflow_run.artifacts_url }}
artifacts_url="https://api.github.com/repos/elastic/cloudbeat/actions/runs/2664140014/artifacts"

gh api "$artifacts_url" -q '.artifacts[] | [.name, .archive_download_url] | @tsv' | while read artifact
do
  IFS=$'\t' read name url <<< "$artifact"
  gh api $url > "$name.zip"
  unzip -d "../tests/allure/results" "$name.zip"
done