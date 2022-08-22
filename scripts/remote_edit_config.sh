#!/bin/bash
# set -x
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

find_config_file() {
  POD=$1
  PREFIX="/usr/share/elastic-agent/data"
  PATH_PARTS=("elastic-agent-" "install" "cloudbeat-" "cloudbeat.yml")
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

TMP_LOCAL="/tmp/cloudbeat.yml"

PODS=$(kubectl -n kube-system get pod -l app=elastic-agent -o name)
for P in $PODS; do
  POD=$(echo "$P" | cut -d '/' -f 2)
  CONFIG_FILEPATH="$(find_config_file $POD)"
  if [ -z "$CONFIG_FILEPATH" ]
  then
    echo "could not find remote config file"
    exit 1
  fi

  kubectl -n kube-system cp "$POD":"$CONFIG_FILEPATH" $TMP_LOCAL
  vi $TMP_LOCAL
  kubectl -n kube-system cp $TMP_LOCAL "$POD":"$CONFIG_FILEPATH"
  kubectl -n kube-system exec "$POD" -- chmod go-w "$CONFIG_FILEPATH"
  kubectl -n kube-system exec "$POD" -- chown root:root "$CONFIG_FILEPATH"
  rm $TMP_LOCAL
done
