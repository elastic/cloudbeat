#!/usr/bin/env bash

# create local volume
volume_name="elastic-local"
volume=$(docker volume ls | grep elastic-local | awk '{ print $2 }')
if [ $volume != $volume_name ]; then
  docker volume create $volume_name
fi;


# create registry container unless it already exists
kind_name=$1
reg_name="kind-registry"
reg_port='5001'
if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" != 'true' ]; then
  echo "creating docker registry"
  docker run \
  -d  \
  --mount type=volume,src=$volume_name,dst=/var/lib/registry \
  --restart=always \
  -p "127.0.0.1:${reg_port}:5000" \
  --name "${reg_name}" \
  registry:2
fi

# create a cluster with the local registry enabled in containerd
cluster=$(kind get clusters | grep $kind_name)
if [[ $cluster == "" ]] || [[ $cluster == "No kind clusters found." ]] ; then
  echo "creating cluster"
  kind create cluster --config=deploy/k8s/kind/$kind_name.yml --wait 30s > /dev/null 2>&1
fi

# connect the registry to the cluster network if not already connected
if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${reg_name}")" = 'null' ]; then
  echo "connecting registry to cluster network"
  docker network connect "kind" "${reg_name}"
fi

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | kubectl apply -f - > /dev/null
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

cat << EOF
Cluster $kind_name is ready
Usage:
  $ kubectl config use-context $kind_name

Local registry is available on port ${reg_port}
  $ docker tag {IMAGE} localhost:${reg_port}/{IMAGE}
  $ docker push localhost:${reg_port}/{IMAGE}
Later to use from within a cluster use the image "localhost:${reg_port}/{IMAGE}" 
EOF