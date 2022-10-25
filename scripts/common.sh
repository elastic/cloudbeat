#!/bin/bash

# Based on the kubeconfig on the running host,
# the function remotely searches for a file with a given pattern on a given folder
# $1 is the pod on which it should run
# $2 is the folder on which it should search
# $3 is a regex pattern of the file which you want to find
# The function returns the file name
find_in_folder() {
  POD=$1
  FOLDER=$2
  PATTERN=$3
  RES=$(kubectl -n kube-system exec $POD -- ls -1 $FOLDER)
  if [ -z "$RES" ]
  then
    return
  fi

  RES=$(echo "$RES" | grep $PATTERN)
  if [ -z "$RES" ]
  then
    return
  fi

  echo "${RES}"
}

# Based on the kubeconfig on the running host,
# the function remotely searches for a file with a given pattern on cloudbeat installation folder
# /usr/share/elastic-agent/data/elastic-agent-*/install/cloudbeat-*/
# $1 is the pod on which it should run
# $2 is a regex pattern of the file which you want to find
# The function returns the file full path
find_in_cloudbeat_folder() {
  POD=$1

  PREFIX="/usr/share/elastic-agent/data"
  PATH_PARTS=("elastic-agent-" "install" "cloudbeat-" "$2")
  for NEXT in ${PATH_PARTS[@]}; do
    FOUND=$(find_in_folder $POD $PREFIX $NEXT)
    if [ -z "$FOUND" ]
    then
      return
    fi
    PREFIX="${PREFIX}/${FOUND}"
  done

  echo $PREFIX
}

find_cloudbeat_config() {
  find_in_cloudbeat_folder $1 "cloudbeat.yml"
}

find_cloudbeat_binary() {
  find_in_cloudbeat_folder $1 "cloudbeat$"
}
