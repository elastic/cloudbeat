package compliance.policy.file.data_adapter

import future.keywords.if

is_filesystem if {
	input.type == "file"
}

filename := file_name if {
	is_filesystem
	file_name := input.resource.name
}

filemode := file_mode if {
	is_filesystem
	file_mode := input.resource.mode
}

file_path := path if {
	is_filesystem
	path := input.resource.path
}

owner_user := owner if {
	is_filesystem
	owner := input.resource.owner
}

owner_group := group if {
	is_filesystem
	group := input.resource.group
}
