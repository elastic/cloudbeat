package compliance.cis_eks.rules.cis_3_2_10

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Verify that the --rotate-certificates argument is not present, or is set to true.

default rule_evaluation = true

process_args := cis_eks.data_adapter.process_args

rule_evaluation = false {
	common.contains_key_with_value(process_args, "--rotate-certificates", "false")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation = false {
	not process_args["--rotate-certificates"]
	data_adapter.process_config.config.rotateCertificates == false
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
