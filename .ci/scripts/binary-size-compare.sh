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
if [ $# -ne 3 ]; then
    echo "Usage: $0 <pr_binary_path> <target_binary_path> <threshold_mib>" >&2
    echo "" >&2
    echo "Example: $0 /tmp/cloudbeat-pr /tmp/cloudbeat-target 5" >&2
    exit 1
fi

PR_BINARY_PATH="$1"
TARGET_BINARY_PATH="$2"
SIZE_THRESHOLD_MIB="$3"

# Verify binary files exist
if [ ! -f "$PR_BINARY_PATH" ]; then
    echo "Error: PR binary not found at $PR_BINARY_PATH" >&2
    exit 1
fi

if [ ! -f "$TARGET_BINARY_PATH" ]; then
    echo "Error: Target binary not found at $TARGET_BINARY_PATH" >&2
    exit 1
fi

# Set up output destination
if [ -n "${GITHUB_OUTPUT:-}" ]; then
    OUTPUT_FILE="$GITHUB_OUTPUT"
    LOCAL_MODE=false
else
    OUTPUT_FILE=$(mktemp)
    LOCAL_MODE=true
    echo "Running in local mode - results will be displayed at the end"
    echo "Output file: $OUTPUT_FILE"
    echo ""
fi

# Get file sizes
PR_SIZE=$(stat -c%s "$PR_BINARY_PATH")
TARGET_SIZE=$(stat -c%s "$TARGET_BINARY_PATH")

echo "PR binary size: $PR_SIZE bytes"
echo "Target binary size: $TARGET_SIZE bytes"

# Calculate human-readable sizes in MiB (1024^2)
PR_SIZE_MIB=$(echo "scale=2; $PR_SIZE/1024/1024" | bc)
TARGET_SIZE_MIB=$(echo "scale=2; $TARGET_SIZE/1024/1024" | bc)

echo "PR binary size: ${PR_SIZE_MIB} MiB"
echo "Target binary size: ${TARGET_SIZE_MIB} MiB"

# Calculate size difference and percentage
if [ $TARGET_SIZE -eq 0 ]; then
    echo "Error: Target branch binary size is 0" >&2
    exit 1
fi

SIZE_DIFF=$((PR_SIZE - TARGET_SIZE))
SIZE_DIFF_MIB=$(echo "scale=2; $SIZE_DIFF/1024/1024" | bc)
PERCENTAGE=$(echo "scale=2; $SIZE_DIFF * 100 / $TARGET_SIZE" | bc)

echo "Size difference: $SIZE_DIFF bytes (${SIZE_DIFF_MIB} MiB)"
echo "Percentage change: ${PERCENTAGE}%"

# Store values for later steps (either to GITHUB_OUTPUT or temp file)
echo "pr_size=$PR_SIZE" >> "$OUTPUT_FILE"
echo "target_size=$TARGET_SIZE" >> "$OUTPUT_FILE"
echo "pr_size_mib=$PR_SIZE_MIB" >> "$OUTPUT_FILE"
echo "target_size_mib=$TARGET_SIZE_MIB" >> "$OUTPUT_FILE"
echo "size_diff=$SIZE_DIFF" >> "$OUTPUT_FILE"
echo "size_diff_mib=$SIZE_DIFF_MIB" >> "$OUTPUT_FILE"
echo "percentage=$PERCENTAGE" >> "$OUTPUT_FILE"

# Check if size increase exceeds threshold
if (( $(echo "$SIZE_DIFF_MIB > $SIZE_THRESHOLD_MIB" | bc -l) )); then
    echo "threshold_exceeded=1" >> "$OUTPUT_FILE"
    echo "significant_increase=1" >> "$OUTPUT_FILE"
    THRESHOLD_EXCEEDED=true
else
    echo "threshold_exceeded=0" >> "$OUTPUT_FILE"
    echo "significant_increase=0" >> "$OUTPUT_FILE"
    THRESHOLD_EXCEEDED=false
fi

# Also store absolute percentage for display
ABS_PERCENTAGE=$(echo "$PERCENTAGE" | sed 's/-//')
echo "abs_percentage=$ABS_PERCENTAGE" >> "$OUTPUT_FILE"

# For local mode, display the results in a nice format
if [ "$LOCAL_MODE" = true ]; then
    echo ""
    echo "=============================================="
    echo "📊 Binary Size Comparison Report"
    echo "=============================================="
    echo ""
    echo "| Branch    | Size (MiB) | Size (bytes) |"
    echo "|-----------|------------|--------------|"
    printf "| PR Branch | %10s | %12s |\n" "${PR_SIZE_MIB}" "$PR_SIZE"
    printf "| Target    | %10s | %12s |\n" "${TARGET_SIZE_MIB}" "$TARGET_SIZE"
    printf "| Difference| %10s | %12s |\n" "${SIZE_DIFF_MIB}" "$SIZE_DIFF"
    echo ""
    echo "Size Change: ${PERCENTAGE}% | Absolute Change: ${SIZE_DIFF_MIB} MiB"
    echo ""
    
    if [ "$THRESHOLD_EXCEEDED" = true ]; then
        echo "⚠️  WARNING: Binary size increased by ${SIZE_DIFF_MIB} MiB, which exceeds the ${SIZE_THRESHOLD_MIB} MiB threshold!"
        echo "   This change would cause the CI workflow to fail."
        echo "   Consider optimizing your changes to reduce binary size impact."
    else
        if (( $(echo "$SIZE_DIFF_MIB < 0" | bc -l) )); then
            echo "✅ Great! Binary size decreased by $(echo "$SIZE_DIFF_MIB" | sed 's/-//')% MiB"
        else
            echo "✅ Binary size change is within acceptable limits."
        fi
    fi
    echo ""
    echo "Results stored in: $OUTPUT_FILE"
    echo ""
    echo "Environment variables available:"
    echo "--------------------------------"
    cat "$OUTPUT_FILE"
    echo ""
    
    # Clean up temp file
    rm -f "$OUTPUT_FILE"
fi

# Exit with appropriate code for CI usage
if [ "$THRESHOLD_EXCEEDED" = true ]; then
    exit 1
else
    exit 0
fi