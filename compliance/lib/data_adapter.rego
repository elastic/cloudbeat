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

is_process {
	input.type == "api_server"
}

# split the process args string into an array
command_args = args {
	is_process
	args = split(input.command, " ")
}
