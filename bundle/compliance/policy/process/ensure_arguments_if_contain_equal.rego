package compliance.policy.process.ensure_arguments_if_contain_equal

import data.benchmark_data_adapter
import data.compliance.lib.assert
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.common as process_common
import data.compliance.policy.process.data_adapter

process_args := data_adapter.process_args(benchmark_data_adapter.process_args_seperator)

default rule_evaluation = false

rule_evaluation {
	process_common.contains_key_with_value(process_args, "--service-account-lookup", "true")
} else {
	assert.is_false(process_common.contains_key(process_args, "--service-account-lookup"))
}

finding = result {
	data_adapter.is_kube_apiserver

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"process_args": process_args},
	)
}
