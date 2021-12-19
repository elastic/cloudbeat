package compliance.lib.data_adapter

import data.compliance.lib.common

is_filesystem {
	input.type == "file-system"
}

filename = file_name {
	is_filesystem
	file_name := input.filename
}

filemode = file_mode {
	is_filesystem
	file_mode := input.mode
}

file_path = path {
	is_filesystem
	path := input.path
}

owner_user_id = uid {
	is_filesystem
	uid := input.uid
}

owner_group_id = gid {
	is_filesystem
	gid := input.gid
}

is_process {
	input.type == "process"
}

process_name = name {
	name := process_args_list[0]
}

process_args_list = args_list {
	args_list := split(input.command, " ")
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
