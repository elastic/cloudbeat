#!/bin/bash

# Function to check if the file pattern exists
file_pattern_exists() {
  local pattern=$1
  ls $pattern &> /dev/null
}

# Set the file pattern you want to wait for
file_pattern="$1"

# Wait until the file pattern exists
while ! file_pattern_exists "$file_pattern"; do
  echo "Waiting for files matching pattern: $file_pattern"
  sleep 1
done

# File pattern exists, proceed with your logic here
echo "Files with pattern $file_pattern have appeared!"
