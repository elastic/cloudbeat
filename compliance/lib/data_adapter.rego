package compliance.lib.data_adapter

import data.compliance.lib.common

is_filesystem {
	input.resource.type == "file-system"
}

filename = file_name {
	is_filesystem
	file_name := input.resource.filename
}

filemode = file_mode {
	is_filesystem
	file_mode := input.resource.mode
}

file_path = path {
	is_filesystem
	path := input.resource.path
}

owner_user_id = uid {
	is_filesystem
	uid := input.resource.uid
}

owner_group_id = gid {
	is_filesystem
	gid := input.resource.gid
}

is_process {
	input.resource.type == "process"
}

process_name = name {
	is_process
	name := process_args_list[0]
}

process_args_list = args_list {
	is_process
	args_list := split(input.resource.command, " ")
}

process_args = args {
	args := {arg: value | [arg, value] = common.split_key_value(process_args_list[_])}
}

is_kube_apiserver {
	process_name == "kube-apiserver"
}

is_kube_controller_manger {
	process_name == "kube-controller-manager"
}

is_kube_scheduler {
	process_name == "kube-scheduler"
}

is_etcd {
	process_name == "etcd"
}

is_kubelet {
	process_name == "kubelet"
}

is_kube_api {
	input.resource.type == "kube-api"
}

pod = p {
	input.resource.kind == "Pod"
	p := input.resource
}

containers = c {
	input.resource.kind == "Pod"
	container_types := {"containers", "initContainers"}
	c := pod.spec[container_types[t]]
}
