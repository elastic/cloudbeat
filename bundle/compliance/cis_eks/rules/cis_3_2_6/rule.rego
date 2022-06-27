package compliance.cis_eks.rules.cis_3_2_6

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

default rule_evaluation = false

process_args := cis_eks.data_adapter.process_args

is_kernel_protected {
	common.contains_key_with_value(process_args, "--protect-kernel-defaults", "true")
}

# Ensure that the --protect-kernel-defaults argument is set to true
rule_evaluation {
	is_kernel_protected
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
# Ensure that the protectKernelDefaults argument is set to true
rule_evaluation {
	not process_args["--protect-kernel-defaults"]
	data_adapter.process_config.config.protectKernelDefaults
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
