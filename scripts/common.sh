#!/bin/bash

exec_pod() {
  pod=$1
  cmd=$2
  kubectl -n kube-system exec $pod -- $cmd
}

cp_to_pod() {
  pod=$1
  source=$2
  dest=$3
  kubectl cp $2 kube-system/$1:$dest
}

get_agents() {
  kubectl -n kube-system get pod -l app=elastic-agent -o name
}

find_target_os() {
  _kubectl_node_info operatingSystem
}

find_target_arch() {
  _kubectl_node_info architecture
}

is_eks() {
  _kubectl_node_info kubeletVersion | grep "eks"
}

get_agent_sha() {
  out=$(exec_pod $1 "elastic-agent version --yaml --binary-only")
  echo $out | cut -d ":" -f4 | awk '{$1=$1};1'|  awk '{ print substr($0, 0, 6) }'
}

_kubectl_node_info() {
 kubectl get node -o go-template="{{(index .items 0 ).status.nodeInfo.$1}}" 
}