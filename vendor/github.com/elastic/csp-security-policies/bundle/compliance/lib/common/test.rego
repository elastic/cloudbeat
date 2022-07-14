package compliance.lib.common

import data.compliance.lib.assert

test_calculate_result_rule_evaluation_false {
	rule_evaluation := false
	calculate_result(rule_evaluation) == "failed"
}

test_calculate_result_rule_evaluation_true {
	rule_evaluation := true
	calculate_result(rule_evaluation) == "passed"
}

test_file_ownership_match_match {
	user := "root"
	group := "root"
	requierd_user := "root"
	requierd_group := "root"
	file_ownership_match(user, group, requierd_user, requierd_group)
}

test_file_ownership_match_user_mismatch {
	user := "owner"
	group := "root"
	requierd_user := "root"
	requierd_group := "root"
	assert.is_false(file_ownership_match(user, group, requierd_user, requierd_group))
}

test_file_ownership_match_gid_mismatch {
	user := "root"
	group := "owner"
	requierd_user := "root"
	requierd_group := "root"
	assert.is_false(file_ownership_match(user, group, requierd_user, requierd_group))
}

test_file_ownership_match_user_gid_mismatch {
	user := "owner"
	group := "owner"
	requierd_user := "root"
	requierd_group := "root"
	assert.is_false(file_ownership_match(user, group, requierd_user, requierd_group))
}

test_file_permission_match {
	users := [0, 1, 2, 3, 4, 5, 6, 7]
	groups := [0, 1, 2, 3, 4, 5, 6, 7]
	others := [0, 1, 2, 3, 4, 5, 6, 7]

	results := {file_permission_match(filemode, 7, 7, 7) | filemode := sprintf("%d%d%d", [users[u], groups[g], others[o]])}
	assert.all_true(results)
}

test_file_permission_match_octal {
	users := [0, 1, 2, 3, 4, 5, 6, 7]
	groups := [0, 1, 2, 3, 4, 5, 6, 7]
	others := [0, 1, 2, 3, 4, 5, 6, 7]

	# some file premission are in ocatl notation (prefix 0)
	prefix_results := {file_permission_match(filemode, 7, 7, 7) | filemode := sprintf("0%d%d%d", [users[u], groups[g], others[o]])}
	assert.all_true(prefix_results)
}

test_file_permission_match_user_mismatch {
	max_users := [0, 1, 2, 3, 4, 5, 6]

	filemode := "700"
	results := {file_permission_match(filemode, max_users[u], 7, 7)}
	assert.all_false(results)
}

test_file_permission_match_user_mismatch_ocatl {
	max_users := [0, 1, 2, 3, 4, 5, 6]

	# some file premission are in ocatl notation (prefix 0)
	octal_filemode := "0700"
	octal_results := {file_permission_match(octal_filemode, max_users[u], 7, 7)}
	assert.all_false(octal_results)
}

test_file_permission_match_group_mismatch {
	max_groups := [0, 1, 2, 3, 4, 5, 6]

	filemode := "070"
	results := {file_permission_match(filemode, 7, max_groups[g], 7)}
	assert.all_false(results)
}

test_file_permission_match_group_mismatch_octal {
	max_groups := [0, 1, 2, 3, 4, 5, 6]

	# some file premission are in ocatl notation (prefix 0)
	octal_filemode := "0070"
	octal_results := {file_permission_match(octal_filemode, 7, max_groups[g], 7)}
	assert.all_false(octal_results)
}

test_file_permission_match_other_mismatch {
	max_others := [0, 1, 2, 3, 4, 5, 6]

	filemode := "007"
	results := {file_permission_match(filemode, 7, 7, max_others[o])}
	assert.all_false(results)
}

test_file_permission_match_other_mismatch_octal {
	max_others := [0, 1, 2, 3, 4, 5, 6]

	# some file premission are in ocatl notation (prefix 0)
	octal_filemode := "0007"
	octal_results := {file_permission_match(octal_filemode, 7, 7, max_others[o])}
	assert.all_false(octal_results)
}

test_array_contains {
	array := ["a", "b", "c"]
	key := "c"
	array_contains(array, key)
}

test_array_contains_not_contains {
	array := ["a", "b", "c"]
	key := "d"
	assert.is_false(array_contains(array, key))
}

test_contains_key {
	array := {"a": "aa", "b": "bb"}
	key := "a"
	contains_key(array, key)
}

test_contains_key_not_contains {
	array := {"a": "aa", "b": "bb"}
	key := "c"
	assert.is_false(contains_key(array, key))
}

test_arg_values_contains {
	arguments := {"a": "1,2,3"}
	key := "a"
	value := "2"
	arg_values_contains(arguments, key, value)
}

test_arg_values_contains_missing_key {
	arguments := {"a": "1,2,3"}
	key := "b"
	value := "2"
	assert.is_false(arg_values_contains(arguments, key, value))
}

test_arg_values_contains_missing_value {
	arguments := {"a": "1,2,3"}
	key := "a"
	value := "4"
	assert.is_false(arg_values_contains(arguments, key, value))
}

test_greater_or_equal_greater {
	value := 10
	minimum := 9
	greater_or_equal(value, minimum)
}

test_greater_or_equal_equal {
	value := 10
	minimum := 10
	greater_or_equal(value, minimum)
}

test_greater_or_equal_smaller {
	value := 10
	minimum := 11
	assert.is_false(greater_or_equal(value, minimum))
}

test_duration_gt_greater {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	duration_gt(duration, min_duration)
}

test_duration_gt_equals {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_gt(duration, min_duration))
}

test_duration_gt_smaller {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_gt(duration, min_duration))
}

test_duration_lt_greater {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	assert.is_false(duration_lt(duration, min_duration))
}

test_duration_lt_equals {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_lt(duration, min_duration))
}

test_duration_lt_smaller {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_lt(duration, min_duration)
}

test_duration_gte_greater {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	duration_gte(duration, min_duration)
}

test_duration_gte_equals {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_gte(duration, min_duration)
}

test_duration_gte_smaller {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	assert.is_false(duration_gte(duration, min_duration))
}

test_duration_lte_greater {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	assert.is_false(duration_lte(duration, min_duration))
}

test_duration_lte_equals {
	duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_lte(duration, min_duration)
}

test_duration_lte_smaller {
	duration := "10h30m15s9ns" # 10 hours 30 minutes 15 seconds 9 nano-seconds
	min_duration := "10h30m15s10ns" # 10 hours 30 minutes 15 seconds 10 nano-seconds
	duration_lte(duration, min_duration)
}

test_file_in_path {
	path := "/path/to/file/"
	file_path := "/path/to/file/my_file.txt"
	file_in_path(path, file_path)
}

test_file_in_path_recursive {
	path := "/path/to/file/"
	file_path := "/path/to/file/dir1/dir2/dir3/my_file.txt"
	file_in_path(path, file_path)
}

test_file_in_path_not_in_path {
	path := "/path/to/file/"
	file_path := "/path/to/dir/file/my_file.txt"
	assert.is_false(file_in_path(path, file_path))
}

test_split_key_value_with_equality_delimiter {
	key_value_string := "--my-arg-name=some_value=true"
	[arg, value] = split_key_value(key_value_string, "=")
	arg == "--my-arg-name"
	value == "some_value=true"
}

test_split_key_value_with_space_delimiter {
	key_value_string := "--my-arg-name some_value=true"
	[arg, value] = split_key_value(key_value_string, " ")
	arg == "--my-arg-name"
	value == "some_value=true"
}

test_split_key_value_multiple_values_with_equality_delimiter {
	key_value_string := "--my-arg-name=first,second"
	[arg, value] = split_key_value(key_value_string, "=")
	args = {arg: value}
	key = "--my-arg-name"
	arg_values_contains(args, key, "first")
	arg_values_contains(args, key, "second")
}

test_split_key_value_multiple_values_with_space_delimiter {
	key_value_string := "--my-arg-name first,second"
	[arg, value] = split_key_value(key_value_string, " ")
	args = {arg: value}
	key = "--my-arg-name"
	arg_values_contains(args, key, "first")
	arg_values_contains(args, key, "second")
}
