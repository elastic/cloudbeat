package compliance.cis_k8s.rules.cis_1_2_24

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --service-account-lookup argument is set to true (Automated)

# Verify that if the --service-account-lookup argument exists it is set to true.

# evaluate
process_args := cis_k8s.data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	common.contains_key_with_value(process_args, "--service-account-lookup", "true")
} else {
	assert.is_false(common.contains_key(process_args, "--service-account-lookup"))
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
