package compliance.policy.process.ensure_arguments_lte

import data.benchmark_data_adapter
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.data_adapter
import future.keywords.if

process_args := benchmark_data_adapter.process_args

finding(entity, value) := result if {
	data_adapter.is_kube_apiserver

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation(entity, value)),
		{"process_args": process_args},
	)
}

rule_evaluation(entity, value) := false if {
	e := process_args[entity]
	lib_common.duration_lte(e, value)
} else := true
