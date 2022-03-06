#!/bin/bash

# Namespace used for the deployment
NAMESPACE="kube-system"
# Resolve absolute path of the current bash script
SCRIPT_PATH="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
# Resolve and derive absolute path of cloudbeats project from the script path
PROJECT_PATH="$( echo $SCRIPT_PATH | rev | cut -d'/' -f4- | rev )"
# Resolve kind cluster name
# TODO: Maybe not to take the first cluster in list?
KIND_CLUSTER_NAME="$( kind get clusters | head -n 1 )"

# Checking if there is no kind cluster
if [ -z "$KIND_CLUSTER_NAME" ]
then
    echo "ERROR: No kind cluster, please create a kind cluster and re-run the script"
    # Terminate script and indicate error
    exit 1
fi

echo "Start"
cd $PROJECT_PATH;
GOOS=linux go build -v
docker build -t cloudbeat .
kind load docker-image cloudbeat:latest --name $KIND_CLUSTER_NAME
kubectl delete -f deploy/k8s/cloudbeat-ds.yaml
kubectl apply -f deploy/k8s/cloudbeat-ds.yaml

CLOUDBEAT_POD_MAME="$( kubectl get pod --selector="k8s-app=cloudbeat" -n $NAMESPACE | grep cloudbeat | awk '{print $1;}' )"
# Checking if there is no cloudbeat pod
if [ -z "$CLOUDBEAT_POD_MAME" ]
then
    echo "ERROR: No cloudbeat pod found"
    # Terminate script and indicate error
    exit 1
fi

# Waiting for cloudbeat pod to be ready
# TODO: Add some sort of a timeout and exit after a while with error
while [[ $(kubectl get pods $CLOUDBEAT_POD_MAME -n $NAMESPACE -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}') != "True" ]]; do echo "waiting for cloudbeat pod $CLOUDBEAT_POD_MAME" && sleep 1; done

# Output logs from the cloudbeat pod
kubectl logs -f --selector="k8s-app=cloudbeat" -n $NAMESPACE