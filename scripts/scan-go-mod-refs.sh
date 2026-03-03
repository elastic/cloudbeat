#!/usr/bin/env bash
set -euo pipefail

# Usage: scan-go-mod-refs.sh [-f|--fetch] <library_name>
# Prints "ref: <go.mod line>" or "ref: N/A" for main, maintained branch tips, and all tags.
# Example: ./scripts/scan-go-mod-refs.sh "elastic/beats/v7"
# Use -f or --fetch to run 'git fetch origin' before scanning.

FETCH_ORIGIN=false

while [[ $# -gt 0 ]]; do
    case "$1" in
    -f | --fetch)
        FETCH_ORIGIN=true
        shift
        ;;
    *)
        break
        ;;
    esac
done

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 [-f|--fetch] <library_name>" >&2
    echo "Example: $0 --fetch \"elastic/beats/v7\"" >&2
    exit 1
fi

LIBRARY_NAME="$1"

if [[ "$FETCH_ORIGIN" == true ]]; then
    echo "Fetching origin..." >&2
    git fetch origin
fi

# Get maintained minor branches (e.g. 8.19, 9.2, 9.3)
MINORS=$(curl -sL "https://elastic-release-api.s3.us-west-2.amazonaws.com/public/future-releases.json" | jq -r '.releases[].version' | sort -V | awk -F. '{key=$1"."$2; if(key!=p){if(p) print p; p=key}} END{if(p) print p}')

# Build ordered list of refs: main, then each minor's branch + tags
REFS=()

# main branch
if git ls-remote --exit-code --heads origin main &>/dev/null; then
    REFS+=("origin/main")
fi

# For each maintained minor: branch tip, then all tags vX.Y.*
while IFS= read -r minor; do
    [[ -z "$minor" ]] && continue
    if git ls-remote --exit-code --heads origin "$minor" &>/dev/null; then
        REFS+=("origin/$minor")
    fi
    while IFS= read -r tag; do
        [[ -z "$tag" ]] && continue
        REFS+=("$tag")
    done < <(git tag -l "v${minor}.*" 2>/dev/null | sort -V || true)
done <<<"$MINORS"

# Print go.mod line for each ref in order
show_ref() {
    local ref="$1"
    local line
    if line=$(git show "$ref":go.mod 2>/dev/null | grep -F "$LIBRARY_NAME"); then
        echo "$ref: $line"
    else
        echo "$ref: N/A"
    fi
}

for ref in "${REFS[@]}"; do
    show_ref "$ref"
done
