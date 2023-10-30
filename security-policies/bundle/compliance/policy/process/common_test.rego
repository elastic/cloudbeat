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
