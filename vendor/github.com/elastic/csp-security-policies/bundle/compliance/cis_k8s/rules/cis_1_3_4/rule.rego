package compliance.cis_k8s.rules.cis_1_3_4

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --service-account-private-key-file argument is set as appropriate (Automated)
finding = result {
	# filter
	data_adapter.is_kube_controller_manager

	# evaluate
	process_args := cis_k8s.data_adapter.process_args
	rule_evaluation := common.contains_key(process_args, "--service-account-private-key-file")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}
