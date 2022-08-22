package compliance.policy.kube_api.data_adapter

import future.keywords.in

is_kube_api {
	input.type == "k8s_object"
}

is_cluster_roles {
	is_kube_api
	input.resource.kind in {"Role", "ClusterRole"}
}

cluster_roles := roles {
	is_cluster_roles
	roles = input.resource
}

service_account := account {
	input.resource.kind == "ServiceAccount"
	account = input.resource
}

is_kube_node {
	is_kube_api
	input.resource.kind == "Node"
}

pod = p {
	input.resource.kind == "Pod"
	p := input.resource
}

is_service_account_or_pod = pod

is_service_account_or_pod = service_account

containers = c {
	input.resource.kind == "Pod"
	container_types := {"containers", "initContainers"}
	c := pod.spec[container_types[t]]
}

status = input.resource.status
