#!/bin/bash
# Retry wrapper script for commands that may fail due to transient errors
#
# Usage: retry.sh [max_retries] [initial_delay] <command>
#   max_retries: Number of retry attempts (default: 5)
#   initial_delay: Initial delay in seconds between retries (default: 2)
#   command: The command to execute

set -euo pipefail

# Capture the original working directory to ensure commands run from the correct location
original_dir="$(pwd)"

max_retries="${1:-5}"
retry_delay="${2:-2}"
shift 2 || true # Remove first two args if they exist, otherwise shift nothing
command="$*"

if [ -z "$command" ]; then
    echo "Error: No command provided"
    echo "Usage: retry.sh [max_retries] [initial_delay] <command>"
    exit 1
fi

attempt=1
current_delay=$retry_delay

while [ $attempt -le "$max_retries" ]; do
    # Ensure we're in the original directory for each attempt
    cd "$original_dir"

    echo "Attempt $attempt/$max_retries: Executing command..."
    echo "Command: $command"
    echo "Working directory: $(pwd)"

    eval "$command"
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
        echo "Command succeeded on attempt $attempt"
        exit 0
    fi

    if [ $attempt -lt "$max_retries" ]; then
        echo "Command failed (exit code: $exit_code). Retrying in ${current_delay}s..."
        sleep "$current_delay"
        current_delay=$((current_delay * 2)) # Exponential backoff
        attempt=$((attempt + 1))
    else
        echo "Command failed after $max_retries attempts with exit code: $exit_code"
        exit $exit_code
    fi
done
