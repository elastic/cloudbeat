package compliance.policy.process.ensure_arguments_not_contain_multiple_values

import data.compliance.lib.common as lib_common
import data.compliance.policy.process.common as process_common
import data.compliance.policy.process.data_adapter

default rule_evaluation = true

rule_evaluation = false {
	# Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual)
	not process_common.arg_values_contains(data_adapter.process_args, "--enable-admission-plugins", "SecurityContextDeny")
	not process_common.arg_values_contains(data_adapter.process_args, "--enable-admission-plugins", "PodSecurityPolicy")
}

finding = result {
	data_adapter.is_kube_apiserver

	# set result
	result := lib_common.generate_result_without_expected(
		lib_common.calculate_result(rule_evaluation),
		{"process_args": data_adapter.process_args},
	)
}
