package compliance.policy.file.common

import data.compliance.lib.assert
import future.keywords.if

file_ownership_match(user, group, required_user, required_group) if {
	user == required_user
	group == required_group
} else := false

file_permission_match(filemode, user, group, other) if {
	permissions = parse_permission(filemode)

	# filemode format {user}{group}{other} e.g. 644
	check_permissions(permissions, [user, group, other])
} else := false

file_permission_match_exact(filemode, user, group, other) if {
	permissions = parse_permission(filemode)

	# filemode format {user}{group}{other} e.g. 644
	permissions == [user, group, other]
} else := false

# return a list of file premission [user, group, other]
# cast to numbers
parse_permission(filemode) := [to_number(p) | p := split(filemode, "")[_]]

check_permissions(permissions, max_permissions) if {
	assert.all_true([r | some p; r = bits.and(permissions[p], bits.negate(max_permissions[p])) == 0])
} else := false

# check if file is in path
file_in_path(path, file_path) if {
	closed_path := concat("", [file_path, "/"]) # make sure last dir name is closed by "/"
	contains(closed_path, path)
} else := false
