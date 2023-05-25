#!/bin/bash

# Check if JSON file path and Allure report URL are provided as arguments
if [ $# -ne 2 ]; then
  echo "Error: history-trend.json file path and/or Allure report URL are missing."
  echo "Usage: $0 <history_trend_path> <allure_report_url>"
  exit 1
fi

# JSON data file path
results="$1"
# Allure report URL
allure_report_url="$2"

echo "Current path: $(pwd)"
echo "Files in $(pwd)"
for file in "$(pwd)"/*; do
    if [[ -f "$file" ]]; then
        echo "$file"
    fi
done

echo "Files in $(pwd)/tests"
for file in "$(pwd)/tests"/*; do
    if [[ -f "$file" ]]; then
        echo "$file"
    fi
done

echo "Files in $(pwd)/tests/allure-report"
for file in "$(pwd)/tests/allure-report"/*; do
    if [[ -f "$file" ]]; then
        echo "$file"
    fi
done

echo "Files in $(pwd)/tests/allure-report/history"
for file in "$(pwd)/tests/allure-report/history"/*; do
    if [[ -f "$file" ]]; then
        echo "$file"
    fi
done
# Check if the JSON file exists
if [ ! -f "$results" ]; then
  echo "Error: JSON file '$results' not found."
  exit 1
fi

# Read JSON data from file
data=$(cat "$results")

# Extract values from JSON using jq
failed=$(echo "$data" | jq -r '.[0].data.failed')
passed=$(echo "$data" | jq -r '.[0].data.passed')
skipped=$(echo "$data" | jq -r '.[0].data.skipped')

# Check if all tests either passed or were skipped
if [ "$failed" -eq 0 ]; then
  summary=":green_heart: All tests either passed or were skipped."
else
  summary=":broken_heart: Some tests failed or were broken."
fi

# Print Summary
echo "### :bar_chart: [Allure Report]($allure_report_url) - $summary"
echo "| Result | Count |"
echo "| :------ | :-----: |"
echo "| 🟥 Failed | $failed |"
echo "| 🟩 Passed | $passed |"
echo "| ⬜ Skipped | $skipped |"
