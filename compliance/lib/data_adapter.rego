package compliance.lib.data_adapter

is_filesystem {
	input.type == "file-system"
}

filename = file_name {
	is_filesystem
	file_name = input.filename
}

filemode = file_mode {
	is_filesystem
	file_mode = input.mode
}

file_path = path {
	is_filesystem
	path = input.path
}

owner_user_id = uid {
	is_filesystem
	uid = input.uid
}

owner_group_id = gid {
	is_filesystem
	gid = input.gid
}

process_args_list = args_list {
	args_list = split(input.command, " ")
}

process_args(args_list) = args {
	args = {arg: value | [arg, value] = split(args_list[_], "=")}
}

is_controller_manager_process {
	input.type == "controller_manager"
}

controller_manager_args = args {
	is_controller_manager_process
	args = process_args(process_args_list)
}

is_api_server_process {
	input.type == "api_server"
}

is_scheduler_process {
	input.type == "scheduler"
}

scheduler_args = args {
	is_scheduler_process
	args = process_args(process_args_list)
}


api_server_command_args = args {
	is_api_server_process
	args = process_args(process_args_list)
}

is_etcd_process {
	input.type == "etcd"
}

etcd_args = args {
	is_etcd_process
	args = process_args(process_args_list)
}
