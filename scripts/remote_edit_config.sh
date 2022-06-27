#!/bin/bash
ARCH=`uname -a | rev | cut -d " " -f 1 | rev` # This should return arm64 on m1 and x86_64 on regular mac
PREFIX="/usr/share/elastic-agent/data/"
SUFFIX="/install/cloudbeat-8.3.0-SNAPSHOT-linux-${ARCH}/cloudbeat.yml"
TMP_LOCAL="/tmp/cloudbeat.yml"

PODS=`kubectl -n kube-system get pod -l app=elastic-agent -o name`
for P in $PODS
do
    POD=`echo $P | cut -d '/' -f 2`
    FOLDER=`kubectl -n kube-system exec $POD -- ls $PREFIX`
    FULL_PATH="${PREFIX}${FOLDER}${SUFFIX}"
    kubectl -n kube-system cp $POD:$FULL_PATH $TMP_LOCAL
    vi $TMP_LOCAL
    kubectl -n kube-system cp $TMP_LOCAL $POD:$FULL_PATH
    kubectl -n kube-system exec $POD -- chmod go-w $FULL_PATH
    kubectl -n kube-system exec $POD -- chown root:root $FULL_PATH
    rm $TMP_LOCAL
done

