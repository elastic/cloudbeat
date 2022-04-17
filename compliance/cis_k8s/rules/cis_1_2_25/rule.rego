package compliance.cis_k8s.rules.cis_1_2_25

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --request-timeout argument is set as appropriate (Automated)

# evaluate
process_args := cis_k8s.data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	value := process_args["--request-timeout"]
	common.duration_gt(value, "60s")
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
