package compliance.policy.file.data_adapter

is_filesystem {
	input.type == "file"
}

filename = file_name {
	is_filesystem
	file_name := input.resource.name
}

filemode = file_mode {
	is_filesystem
	file_mode := input.resource.mode
}

file_path = path {
	is_filesystem
	path := input.resource.path
}

owner_user = owner {
	is_filesystem
	owner := input.resource.owner
}

owner_group = group {
	is_filesystem
	group := input.resource.group
}
