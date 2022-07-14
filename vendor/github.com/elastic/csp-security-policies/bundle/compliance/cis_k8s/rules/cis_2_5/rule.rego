package compliance.cis_k8s.rules.cis_2_5

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --peer-client-cert-auth argument is set to true (Automated)
finding = result {
	# filter
	data_adapter.is_etcd

	# evaluate
	process_args := cis_k8s.data_adapter.process_args
	rule_evaluation := common.contains_key_with_value(process_args, "--peer-client-cert-auth", "true")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}
