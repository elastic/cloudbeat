package compliance.policy.process.common

import future.keywords.if
import future.keywords.in

# checks if argument contains value (argument format is csv)
arg_values_contains(arguments, key, value) if {
	argument := arguments[key]
	values := split(argument, ",")
	value in values
} else := false

# splits key value string by first occurrence of =
split_key_value(key_value_string, delimiter) := [key, value] if {
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
