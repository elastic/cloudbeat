package compliance.policy.process.common

import data.compliance.lib.assert

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
