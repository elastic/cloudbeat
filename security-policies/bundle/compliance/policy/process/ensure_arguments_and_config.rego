package compliance.policy.process.ensure_arguments_and_config

import future.keywords.if
import future.keywords.in

import data.benchmark_data_adapter
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.data_adapter

process_args := benchmark_data_adapter.process_args

finding(rule_evaluation) := result if {
	data_adapter.is_kubelet

	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{
			"process_args": process_args,
			"process_config": data_adapter.process_config,
		},
	)
}

process_contains_key_with_value(entity, value) if {
	lib_common.contains_key_with_value(process_args, entity, value)
}

not_process_contains_key_with_value(entity, value) if {
	not lib_common.contains_key_with_value(process_args, entity, value)
}

not_process_arg_variable(entity, variable) if {
	not process_args[entity]
	get_from_config(variable)
}

process_arg_variable(entity, variable) if {
	process_args[entity]
	get_from_config(variable)
}

not_process_contains_variable(entity, value, variable) if {
	not contains(process_args[entity], value)
	get_from_config(variable)
}

process_arg_not_key_value(entity, key, value) if {
	process_args[entity]
	not_process_contains_key_with_value(key, value)
}

process_contains_key(entity) if {
	entity in object.keys(process_args)
}

not_process_key_comparison(entity, variable, value) if {
	not process_contains_key(entity)
	get_from_config(variable) == value
}

not_process_arg_comparison(entity, variable, value) if {
	not process_args[entity]
	get_from_config(variable) == value
}

process_arg_multi(f_entity, s_entity) if {
	process_args[f_entity]
	process_args[s_entity]
}

process_variable_multi(f_variable, s_variable) if {
	get_from_config(f_variable)
	get_from_config(s_variable)
}

process_filter_variable_multi_comparison(f_variable, s_variable, value) if {
	get_from_config(f_variable)
	not get_from_config(s_variable) == value
}

get_from_config(path) := r if {
	# TODO: object.get needs to be provided with a default value to assign
	# Decided to assign undefined string for non-existing process flag values
	# Another option was to assign a non-string undefined value via "hack" (assign non-existent variable)
	# Did not see a direct option to assign undefined values in rego as of current
	# Rego also has a unique behavior with undefined values that I wanted to avoid
	# Assuming that process flags won't have undefined string values and will be empty or non-existent
	r := object.get(data_adapter.process_config.config, path, "undefined")

	# TODO: This is a "hack" to avoid returning undefined values and recognize when there is no value
	r != "undefined"
}
