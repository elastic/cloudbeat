package compliance.lib.common

metadata = {"opa_version": opa_version}

# get OPA version
opa_version := opa.runtime().version

# set the rule result
calculate_result(evaluation) = "passed" {
	evaluation
} else = "failed" {
	true
}

file_ownership_match(uid, gid, requierd_uid, requierd_gid) {
	uid == requierd_uid
	gid == requierd_gid
} else = false {
	true
}

# todo: compare performance of regex alternatives
file_permission_match(filemode, user, group, other) {
	pattern = sprintf("0?[0-%d][0-%d][0-%d]", [user, group, other])
	regex.match(pattern, filemode)
} else = false {
	true
}

array_contains(array, key) {
	contains(array[_], key)
} else = false {
	true
}

# gets argument's value
get_arg_value(arguments, key) = value {
	contains(arguments[i], key)
	argument := arguments[i]
	[_, value] := split(argument, "=")
}

# checks if argument contains value (argument format is csv)
arg_values_contains(arguments, key, value) {
	argument := get_arg_value(arguments, key)
	values := split(argument, ",")
	value = values[_]
} else = false {
	true
}

# checks if a argument is set to greater value then minimum
arg_at_least(arguments, key, minimum) {
	value := get_arg_value(arguments, key)
	to_number(value) >= minimum
} else = false {
	true
}

# check if file is in path
file_in_path(path, file_path) {
	closed_path := concat("", [file_path, "/"]) # make sure last dir name is closed by "/"
	contains(closed_path, path)
}
