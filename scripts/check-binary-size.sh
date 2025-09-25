#!/bin/bash

# Local binary size checker
# This script helps developers check binary size changes locally before pushing

set -e

# Configuration
THRESHOLD=${SIZE_THRESHOLD:-10}
TEMP_DIR=$(mktemp -d)
CURRENT_BRANCH=$(git branch --show-current)

echo "🔍 Binary Size Checker"
echo "======================="
echo "Current branch: $CURRENT_BRANCH"
echo "Size threshold: $THRESHOLD%"
echo ""

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# Build current branch binary
echo "📦 Building current branch binary..."
if command -v mage &> /dev/null && mage build; then
    cp cloudbeat "$TEMP_DIR/cloudbeat-current"
else
    echo "Using direct go build..."
    go mod vendor
    GOOS=linux CGO_ENABLED=0 go build -tags=grpcnotrace,release -o "$TEMP_DIR/cloudbeat-current" .
fi

# Build main branch binary
echo "📦 Building main branch binary..."
git stash push -m "temp stash for size check" || true
git checkout main
if command -v mage &> /dev/null && mage build; then
    cp cloudbeat "$TEMP_DIR/cloudbeat-main"
else
    echo "Using direct go build..."
    go mod vendor
    GOOS=linux CGO_ENABLED=0 go build -tags=grpcnotrace,release -o "$TEMP_DIR/cloudbeat-main" .
fi

# Return to original branch
git checkout "$CURRENT_BRANCH"
git stash pop || true

# Compare sizes
echo ""
echo "📊 Size Comparison"
echo "=================="

CURRENT_SIZE=$(stat -f%z "$TEMP_DIR/cloudbeat-current" 2>/dev/null || stat -c%s "$TEMP_DIR/cloudbeat-current")
MAIN_SIZE=$(stat -f%z "$TEMP_DIR/cloudbeat-main" 2>/dev/null || stat -c%s "$TEMP_DIR/cloudbeat-main")

CURRENT_SIZE_MB=$(echo "scale=2; $CURRENT_SIZE/1024/1024" | bc)
MAIN_SIZE_MB=$(echo "scale=2; $MAIN_SIZE/1024/1024" | bc)

SIZE_DIFF=$((CURRENT_SIZE - MAIN_SIZE))
SIZE_DIFF_MB=$(echo "scale=2; $SIZE_DIFF/1024/1024" | bc)
PERCENTAGE=$(echo "scale=2; $SIZE_DIFF * 100 / $MAIN_SIZE" | bc)

printf "Current branch: %s bytes (%.2f MB)\n" "$CURRENT_SIZE" "$CURRENT_SIZE_MB"
printf "Main branch:    %s bytes (%.2f MB)\n" "$MAIN_SIZE" "$MAIN_SIZE_MB"
printf "Difference:     %s bytes (%.2f MB)\n" "$SIZE_DIFF" "$SIZE_DIFF_MB"
printf "Change:         %.2f%%\n" "$PERCENTAGE"

echo ""

# Check threshold
if (( $(echo "$PERCENTAGE > $THRESHOLD" | bc -l) )); then
    echo "⚠️  WARNING: Binary size increased by $PERCENTAGE%, exceeding the $THRESHOLD% threshold!"
    echo "   This change would cause the CI workflow to fail."
    echo "   Consider optimizing your changes to reduce binary size impact."
    exit 1
elif (( $(echo "$PERCENTAGE < 0" | bc -l) )); then
    echo "✅ Great! Binary size decreased by $(echo $PERCENTAGE | sed 's/-//')%"
else
    echo "✅ Binary size change is within acceptable limits."
fi

echo ""
echo "💡 Tips to reduce binary size:"
echo "   - Remove unused dependencies"
echo "   - Use build tags to exclude optional features"
echo "   - Review imported packages for size impact"
echo "   - Consider using build optimization flags"