package compliance.cis_k8s.rules.cis_1_2_17

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --insecure-bind-address argument is not set (Automated)
finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# evaluate
	process_args := cis_k8s.data_adapter.process_args
	rule_evaluation := assert.is_false(common.contains_key(process_args, "--insecure-bind-address"))

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}
