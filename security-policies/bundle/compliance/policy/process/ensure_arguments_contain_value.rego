package compliance.policy.process.ensure_arguments_contain_value

import data.benchmark_data_adapter
import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.common as process_common
import data.compliance.policy.process.data_adapter
import future.keywords.if

process_args := benchmark_data_adapter.process_args

finding(rule_evaluation) := result if {
	data_adapter.is_kube_apiserver

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"process_args": process_args},
	)
}

arg_not_contains(entity, value) := assert.is_false(process_common.arg_values_contains(process_args, entity, value))

arg_contains(entity, value) := process_common.arg_values_contains(process_args, entity, value)
