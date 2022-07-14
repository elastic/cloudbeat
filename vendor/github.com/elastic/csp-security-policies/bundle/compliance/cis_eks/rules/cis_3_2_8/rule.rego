package compliance.cis_eks.rules.cis_3_2_8

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --hostname-override argument is not set.

default rule_evaluation = true

process_args := cis_eks.data_adapter.process_args

# Note This setting is not configurable via the Kubelet config file.
rule_evaluation = false {
	common.contains_key(process_args, "--hostname-override")
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
