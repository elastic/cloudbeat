package compliance.policy.process.ensure_arguments_not_contain_multiple_values

import data.benchmark_data_adapter
import data.compliance.lib.common as lib_common
import data.compliance.policy.process.common as process_common
import data.compliance.policy.process.data_adapter
import future.keywords.if

process_args := benchmark_data_adapter.process_args

default rule_evaluation := true

rule_evaluation := false if {
	# Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual)
	not process_common.arg_values_contains(process_args, "--enable-admission-plugins", "SecurityContextDeny")
	not process_common.arg_values_contains(process_args, "--enable-admission-plugins", "PodSecurityPolicy")
}

finding := result if {
	data_adapter.is_kube_apiserver

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"process_args": process_args},
	)
}
