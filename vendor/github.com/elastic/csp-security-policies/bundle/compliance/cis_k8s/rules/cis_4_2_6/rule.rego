package compliance.cis_k8s.rules.cis_4_2_6

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the protectKernelDefaults argument is set to true

default rule_evaluation = false

process_args := cis_k8s.data_adapter.process_args

is_kernel_protected {
	common.contains_key_with_value(process_args, "--protect-kernel-defaults", "true")
}

# Ensure that the --protect-kernel-defaults argument is set to true
rule_evaluation {
	is_kernel_protected
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
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
