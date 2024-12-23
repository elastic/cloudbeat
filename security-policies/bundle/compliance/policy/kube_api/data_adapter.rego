package compliance.policy.kube_api.data_adapter

import future.keywords.if
import future.keywords.in

is_kube_api if {
	input.type == "k8s_object"
}

is_cluster_roles if {
	is_kube_api
	input.resource.kind in {"Role", "ClusterRole"}
}

cluster_roles := roles if {
	is_cluster_roles
	roles = input.resource
}

service_account := account if {
	input.resource.kind == "ServiceAccount"
	account = input.resource
}

is_kube_node if {
	is_kube_api
	input.resource.kind == "Node"
}

is_kube_pod if {
	is_kube_api
	input.resource.kind == "Pod"
}

pod := p if {
	is_kube_pod
	p := input.resource
}

is_service_account_or_pod := pod

is_service_account_or_pod := service_account

containers := c if {
	is_kube_pod
	c := {
		"app_containers": object.get(pod.spec, "containers", {}),
		"init_containers": object.get(pod.spec, "initContainers", {}),
		"ephemeral_containers": object.get(pod.spec, "ephemeralContainers", {}),
	}
}

status := input.resource.status
