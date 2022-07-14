package compliance.cis_k8s.rules.cis_1_2_13

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

default rule_evaluation = true

process_args := cis_k8s.data_adapter.process_args

# evaluate
rule_evaluation = false {
	# Ensure that the admission control plugin SecurityContextDeny is set if PodSecurityPolicy is not used (Manual)
	not common.arg_values_contains(process_args, "--enable-admission-plugins", "SecurityContextDeny")
	not common.arg_values_contains(process_args, "--enable-admission-plugins", "PodSecurityPolicy")
}

finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}
