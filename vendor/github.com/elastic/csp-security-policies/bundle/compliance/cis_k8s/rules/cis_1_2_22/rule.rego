package compliance.cis_k8s.rules.cis_1_2_22

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated)

# evaluate
process_args := cis_k8s.data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	value := process_args["--audit-log-maxsize"]
	common.greater_or_equal(value, 100)
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
