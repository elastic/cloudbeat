package compliance.policy.process.ensure_arguments_if_contain_equal

import future.keywords.if
import future.keywords.in

import data.benchmark_data_adapter
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.data_adapter

process_args := benchmark_data_adapter.process_args

default rule_evaluation := false

rule_evaluation if {
	lib_common.contains_key_with_value(process_args, "--service-account-lookup", "true")
} else if {
	not "--service-account-lookup" in object.keys(process_args)
}

finding := result if {
	data_adapter.is_kube_apiserver

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"process_args": process_args},
	)
}
