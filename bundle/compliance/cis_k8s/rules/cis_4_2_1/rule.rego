package compliance.cis_k8s.rules.cis_4_2_1

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

default rule_evaluation = false

process_args := cis_k8s.data_adapter.process_args

rule_evaluation {
	common.contains_key_with_value(process_args, "--anonymous-auth", "false")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for authentication:anonymous: enabled set to false.
rule_evaluation {
	not process_args["--anonymous-auth"]
	assert.is_false(data_adapter.process_config.config.authentication.anonymous.enabled)
}

# Ensure that the --anonymous-auth argument is set to false (Automated)
finding = result {
	# filter
	data_adapter.is_kubelet

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {
			"process_args": process_args,
			"process_config": data_adapter.process_config,
		},
	}
}
