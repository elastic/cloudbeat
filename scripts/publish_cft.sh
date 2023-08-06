#!/bin/bash

function usage() {
  cat <<EOF
Usage: $0 <local-file> <file-pattern>
Upload the local file to S3, replacing VERSION and DATE in the file pattern.
EOF
}

LOCAL_FILE=$1
FILEPATTERN=$2
: "${LOCAL_FILE:?$(echo "Missing local file" && usage && exit 1)}"
: "${FILEPATTERN:?$(echo "Missing file pattern" && usage && exit 1)}"

VERSION=$(grep defaultBeatVersion version/version.go | cut -f2 -d "\"")
DATE=$(date +"%Y-%m-%d-%H-%M-%S")
FILEPATTERN=${FILEPATTERN/VERSION/$VERSION}
FILEPATTERN=${FILEPATTERN/DATE/$DATE}

sed -i '' s/CFT_VERSION/$FILEPATTERN/g $LOCAL_FILE

REMOTE_FILE="s3://elastic-cspm-cft/$FILEPATTERN"
echo "Uploading $LOCAL_FILE to $REMOTE_FILE"
aws s3 cp $LOCAL_FILE $REMOTE_FILE