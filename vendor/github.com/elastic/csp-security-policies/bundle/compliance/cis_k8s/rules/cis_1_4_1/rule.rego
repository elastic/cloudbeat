package compliance.cis_k8s.rules.cis_1_4_1

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --profiling argument is set to false (Automated)
finding = result {
	# filter
	data_adapter.is_kube_scheduler

	# evaluate
	process_args := cis_k8s.data_adapter.process_args
	rule_evaluation := common.contains_key_with_value(process_args, "--profiling", "false")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}
