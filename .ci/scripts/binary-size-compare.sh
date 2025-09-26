#!/bin/bash

# Binary size comparison script
# This script compares the sizes of two binary files and reports the results.
# Used by the binary size monitoring workflow and can be run locally for debugging.
#
# Usage: binary-size-compare.sh <pr_binary_path> <target_binary_path> <threshold_mib>
#
# Arguments:
#   pr_binary_path     - Path to the PR branch binary
#   target_binary_path - Path to the target branch binary
#   threshold_mib      - Threshold in MiB for significant size increases
#
# Environment:
#   GITHUB_OUTPUT - If set, outputs variables for GitHub Actions workflow
#                   If not set, displays results locally and stores in temp file

set -euo pipefail

# Check arguments
if [[ $# -lt 2 || $# -gt 3 ]]; then
    echo "Usage: $0 <pr_binary_path> <target_binary_path> <threshold_mib>" >&2
    echo "" >&2
    echo "Example: $0 /tmp/cloudbeat-pr /tmp/cloudbeat-target 5" >&2
    exit 1
fi

pr_binary_path="$1"
target_binary_path="$2"
size_threshold_mib="${3:-5}"

# Verify binary files exist
if [ ! -f "$pr_binary_path" ]; then
    echo "Error: PR binary not found at $pr_binary_path" >&2
    exit 1
fi

if [ ! -f "$target_binary_path" ]; then
    echo "Error: Target binary not found at $target_binary_path" >&2
    exit 1
fi

# Set up output destination
if [ -n "${GITHUB_OUTPUT:-}" ]; then
    output_file="$GITHUB_OUTPUT"
else
    output_file=$(mktemp)
fi

# Get file sizes
pr_size=$(stat -c%s "$pr_binary_path")
target_size=$(stat -c%s "$target_binary_path")

echo "PR binary size: $pr_size bytes"
echo "Target binary size: $target_size bytes"

# Calculate human-readable sizes in MiB (1024^2)
pr_size_mib=$(echo "scale=2; $pr_size/1024/1024" | bc)
target_size_mib=$(echo "scale=2; $target_size/1024/1024" | bc)

echo "PR binary size: ${pr_size_mib} MiB"
echo "Target binary size: ${target_size_mib} MiB"

# Calculate size difference and percentage
if [ "$target_size" -eq 0 ]; then
    echo "Error: Target branch binary size is 0" >&2
    exit 1
fi

size_diff=$((pr_size - target_size))
size_diff_mib=$(echo "scale=2; $size_diff/1024/1024" | bc)
percentage=$(echo "scale=2; $size_diff * 100 / $target_size" | bc)

echo "Size difference: $size_diff bytes (${size_diff_mib} MiB)"
echo "Percentage change: ${percentage}%"

# Store values for later steps (either to GITHUB_OUTPUT or temp file)
{
    echo "pr_size=$pr_size"
    echo "target_size=$target_size"
    echo "pr_size_mib=$pr_size_mib"
    echo "target_size_mib=$target_size_mib"
    echo "size_diff=$size_diff"
    echo "size_diff_mib=$size_diff_mib"
    echo "percentage=$percentage"
} >>"$output_file"

# Check if size increase exceeds threshold
if (($(echo "$size_diff_mib > $size_threshold_mib" | bc -l))); then
    echo "threshold_exceeded=1" >>"$output_file"
else
    echo "threshold_exceeded=0" >>"$output_file"
fi

# Also store absolute percentage for display
abs_percentage=${percentage/-/}
echo "abs_percentage=$abs_percentage" >>"$output_file"

echo ""
cat "$output_file"
