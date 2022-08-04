package compliance.policy.file.common

import data.compliance.lib.assert

file_ownership_match(user, group, required_user, required_group) {
	user == required_user
	group == required_group
} else = false {
	true
}

file_permission_match(filemode, user, group, other) {
	permissions = parse_permission(filemode)

	# filemode format {user}{group}{other} e.g. 644
	check_permissions(permissions, [user, group, other])
} else = false {
	true
}

# in some os filemodes starts with 0 to indicate that the value is Octal (base 8)
# remove prefix if needed, and return a list of file premission [user, group, other]
parse_permission(filemode) = permissions {
	# if prefix exist we should start the substring from 1, else 0
	start = count(filemode) - 3

	# remove prefix (if needed) and split
	str_permissions = split(substring(filemode, start, 3), "")

	# cast to numbers
	permissions := [to_number(p) | p = str_permissions[_]]
}

check_permissions(permissions, max_permissions) {
	assert.all_true([r | r = bits.and(permissions[p], bits.negate(max_permissions[p])) == 0])
} else = false {
	true
}

# check if file is in path
file_in_path(path, file_path) {
	closed_path := concat("", [file_path, "/"]) # make sure last dir name is closed by "/"
	contains(closed_path, path)
} else = false {
	true
}
