#!/bin/bash

# Input: version to calculate previous version
VERSION="$1"

# Extract the major and minor versions
MAJOR_VERSION=$(echo "$VERSION" | cut -d'.' -f1)
MINOR_VERSION=$(echo "$VERSION" | cut -d'.' -f2)

# Calculate the previous version (assuming it's always X.(Y-1))
PREVIOUS_VERSION="$MAJOR_VERSION.$((MINOR_VERSION - 1))"

URL="https://snapshots.elastic.co/latest/$PREVIOUS_VERSION.json"

# Use curl to fetch the JSON data
JSON_RESPONSE=$(curl -s "$URL")

# Get latest snapshot version
SNAPSHOT_VERSION=$(echo "$JSON_RESPONSE" | jq -r '.version')

# Check if SNAPSHOT_VERSION is empty
if [ -z "$SNAPSHOT_VERSION" ]; then
    # Log an error message with variable values
    echo "Error: The release version corresponding to $PREVIOUS_VERSION could not be found." >&2
    exit 1
fi

# Split the version into major, minor, and patch parts
IFS='.-' read -ra PARTS <<<"$SNAPSHOT_VERSION"
MAJOR="${PARTS[0]}"
MINOR="${PARTS[1]}"
PATCH="${PARTS[2]}"

# Decrement the patch version by 1
PATCH=$((PATCH - 1))

# Format the previous version
PREVIOUS_VERSION="$MAJOR.$MINOR.$PATCH"

# Output the previous version
echo "$PREVIOUS_VERSION"
