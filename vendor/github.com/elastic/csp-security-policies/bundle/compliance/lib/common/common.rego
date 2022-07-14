package compliance.lib.common

import data.compliance.lib.assert

metadata = {"opa_version": opa_version}

# get OPA version
opa_version := opa.runtime().version

# set the rule result
calculate_result(evaluation) = "passed" {
	evaluation
} else = "failed" {
	true
}

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

array_contains(array, key) {
	contains(array[_], key)
} else = false {
	true
}

contains_key(object, key) {
	object[key]
} else = false {
	true
}

contains_key_with_value(object, key, value) {
	object[key] = value
} else = false {
	true
}

# checks if argument contains value (argument format is csv)
arg_values_contains(arguments, key, value) {
	argument := arguments[key]
	values := split(argument, ",")
	value == values[_]
} else = false {
	true
}

# checks if a value is greater or equals to a minimum value
greater_or_equal(value, minimum) {
	to_number(value) >= minimum
} else = false {
	true
}

# checks if duration is less than some maximum value
# duration: string (https://pkg.go.dev/time#ParseDuration)
duration_lt(duration, max_duration) {
	duration_ns := time.parse_duration_ns(duration)
	max_duration_ns := time.parse_duration_ns(max_duration)
	duration_ns < max_duration_ns
} else = false {
	true
}

# checks if duration is less than some maximum value
# duration: string (https://pkg.go.dev/time#ParseDuration)
duration_lte(duration, max_duration) {
	duration_ns := time.parse_duration_ns(duration)
	max_duration_ns := time.parse_duration_ns(max_duration)
	duration_ns <= max_duration_ns
} else = false {
	true
}

# checks if duration is greater than some minimum value
# duration: string (https://pkg.go.dev/time#ParseDuration)
duration_gt(duration, min_duration) {
	duration_ns := time.parse_duration_ns(duration)
	min_duration_ns := time.parse_duration_ns(min_duration)
	duration_ns > min_duration_ns
} else = false {
	true
}

# checks if duration is greater or equal to some minimum value
# duration: string (https://pkg.go.dev/time#ParseDuration)
duration_gte(duration, min_duration) {
	duration_ns := time.parse_duration_ns(duration)
	min_duration_ns := time.parse_duration_ns(min_duration)
	duration_ns >= min_duration_ns
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

# splits key value string by first occurrence of =
split_key_value(key_value_string, delimiter) = [key, value] {
	seperator_index := indexof(key_value_string, delimiter)

	# extract key
	key_start_index := 0
	key_length := seperator_index
	key := substring(key_value_string, key_start_index, key_length)

	# extract value
	value_start_index := seperator_index + 1
	value_length := (count(key_value_string) - seperator_index) - 1
	value := substring(key_value_string, value_start_index, value_length)
}

ranges_smaller_than(ranges, value) {
	range := ranges[_]
	range < value
}

ranges_gte(ranges, value) {
	not ranges_smaller_than(ranges, value)
}
