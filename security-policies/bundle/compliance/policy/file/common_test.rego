package compliance.policy.file.common

import data.compliance.lib.assert
import future.keywords.if

test_file_ownership_match_match if {
	user := "root"
	group := "root"
	requierd_user := "root"
	requierd_group := "root"
	file_ownership_match(user, group, requierd_user, requierd_group)
}

test_file_ownership_match_user_mismatch if {
	user := "owner"
	group := "root"
	requierd_user := "root"
	requierd_group := "root"
	assert.is_false(file_ownership_match(user, group, requierd_user, requierd_group))
}

test_file_ownership_match_gid_mismatch if {
	user := "root"
	group := "owner"
	requierd_user := "root"
	requierd_group := "root"
	assert.is_false(file_ownership_match(user, group, requierd_user, requierd_group))
}

test_file_ownership_match_user_gid_mismatch if {
	user := "owner"
	group := "owner"
	requierd_user := "root"
	requierd_group := "root"
	assert.is_false(file_ownership_match(user, group, requierd_user, requierd_group))
}

test_file_permission_match_exact if {
	users := [0, 1, 2, 3, 4, 5, 6, 7]
	groups := [0, 1, 2, 3, 4, 5, 6, 7]
	others := [0, 1, 2, 3, 4, 5, 6, 7]

	results := {file_permission_match_exact(sprintf("%d%d%d", filemode), filemode[0], filemode[1], filemode[2]) | filemode := [users[u], groups[g], others[o]]}
	assert.all_true(results)
}

test_file_permission_match if {
	users := [0, 1, 2, 3, 4, 5, 6, 7]
	groups := [0, 1, 2, 3, 4, 5, 6, 7]
	others := [0, 1, 2, 3, 4, 5, 6, 7]

	results := {file_permission_match(filemode, 7, 7, 7) | filemode := sprintf("%d%d%d", [users[u], groups[g], others[o]])}
	assert.all_true(results)
}

test_file_permission_match_user_mismatch if {
	max_users := [0, 1, 2, 3, 4, 5, 6]

	filemode := "700"
	results := {file_permission_match(filemode, max_users[u], 7, 7)}
	assert.all_false(results)
}

test_file_permission_match_group_mismatch if {
	max_groups := [0, 1, 2, 3, 4, 5, 6]

	filemode := "070"
	results := {file_permission_match(filemode, 7, max_groups[g], 7)}
	assert.all_false(results)
}

test_file_permission_match_other_mismatch if {
	max_others := [0, 1, 2, 3, 4, 5, 6]

	filemode := "007"
	results := {file_permission_match(filemode, 7, 7, max_others[o])}
	assert.all_false(results)
}

test_file_in_path if {
	path := "/path/to/file/"
	file_path := "/path/to/file/my_file.txt"
	file_in_path(path, file_path)
}

test_file_in_path_recursive if {
	path := "/path/to/file/"
	file_path := "/path/to/file/dir1/dir2/dir3/my_file.txt"
	file_in_path(path, file_path)
}

test_file_in_path_not_in_path if {
	path := "/path/to/file/"
	file_path := "/path/to/dir/file/my_file.txt"
	assert.is_false(file_in_path(path, file_path))
}
