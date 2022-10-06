#!/bin/bash

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
