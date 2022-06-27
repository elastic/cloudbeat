package compliance.cis_k8s.rules.cis_1_2_14

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the admission control plugin ServiceAccount is set (Automated)

# evaluate
process_args := cis_k8s.data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	# Verify that the --disable-admission-plugins argument is set to a value that does not includes ServiceAccount.
	process_args["--disable-admission-plugins"]
	not common.arg_values_contains(process_args, "--disable-admission-plugins", "ServiceAccount")
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
