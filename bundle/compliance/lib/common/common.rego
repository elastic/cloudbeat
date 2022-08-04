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

generate_result(evaluation, expected, evidence) = result {
	result := {
		"evaluation": evaluation,
		"expected": expected,
		"evidence": evidence,
	}
}
