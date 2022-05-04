package compliance.cis_eks.rules.cis_3_2_5

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --streaming-connection-idle-timeout argument is not set to 0

default rule_evaluation = true

process_args := cis_eks.data_adapter.process_args

rule_evaluation = false {
	common.contains_key_with_value(process_args, "--streaming-connection-idle-timeout", "0")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation = false {
	not process_args["--streaming-connection-idle-timeout"]
	data_adapter.process_config.config.streamingConnectionIdleTimeout == 0
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
