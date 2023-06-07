#!/bin/bash

declare -A outputs=$1

for var_name in "${!outputs[@]}"; do
  output="${outputs[$var_name]}"
  value=$(terraform output "$output")
  masked_value="::$var_name::"
  echo "::add-mask::$value"
  echo "$var_name=$value" >> $GITHUB_ENV
done
