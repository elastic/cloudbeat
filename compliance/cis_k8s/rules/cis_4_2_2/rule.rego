package compliance.cis_k8s.rules.cis_4_2_2

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)
# If the --authorization-mode argument is present check that it is not set to AlwaysAllow.

default rule_evaluation = false

# evaluate
process_args := cis_k8s.data_adapter.process_args

rule_evaluation {
	is_authorization_allow_all
}

is_authorization_allow_all {
	process_args["--authorization-mode"]
	not common.contains_key_with_value(process_args, "--authorization-mode", "AlwaysAllow")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Checks that the entry for authorization:mode is not set to AlwaysAllow.
rule_evaluation {
	not is_authorization_allow_all
	data_adapter.process_config.config.authorization.mode
	not data_adapter.process_config.config.authorization.mode == "AlwaysAllow"
}

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
