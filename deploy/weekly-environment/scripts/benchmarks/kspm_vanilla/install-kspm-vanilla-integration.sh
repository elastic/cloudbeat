#!/bin/bash

source ../../utils.sh

# This script is used to install a vanilla integration for the KSPM vanilla benchmark.
# It will create a new agent policy, a new vanilla integration and a new vanilla integration manifest file.
# The script requires two arguments:
# 1. Kibana URL
# 2. Kibana password


# using kubectl patch to update manifest file to always pull the latest image

# Replace image in yaml file with my own image

ll manifest.yaml
sed -i "s/^\( *image: *\).*/\1aaa/" manifest.yaml
