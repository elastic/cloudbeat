package compliance.cis_k8s.rules.cis_2_3

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --auto-tls argument is not set to true (Automated)
finding = result {
	# filter
	data_adapter.is_etcd

	# Verify that if the --auto-tls argument exists, it is not set to true.

	# evaluate
	process_args := cis_k8s.data_adapter.process_args
	rule_evaluation := assert.is_false(common.contains_key_with_value(process_args, "--auto-tls", "true"))

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}
