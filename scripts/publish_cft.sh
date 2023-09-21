#!/bin/bash

function usage() {
    cat <<EOF
Usage: $0 <local-file> <remote-file>
Replace the CFT_VERSION placeholder in the local-file.
Upload the local-file to S3 with the remote-file name.

EOF
}

LOCAL_FILE=$1
REMOTE_FILE=$2
: "${LOCAL_FILE:?$(echo "Missing local file" && usage && exit 1)}"
: "${REMOTE_FILE:?$(echo "Missing remote file" && usage && exit 1)}"

sed --in-place'' s/CFT_VERSION/$REMOTE_FILE/g $LOCAL_FILE

S3_FILE="s3://elastic-cspm-cft/$REMOTE_FILE"
echo "Uploading $LOCAL_FILE to $S3_FILE"
aws s3 cp $LOCAL_FILE $S3_FILE
