package compliance.cis_eks.rules.cis_3_2_11

import data.compliance.cis_eks
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Verify that the RotateKubeletServerCertificate argument is set to true

default rule_evaluation = false

process_args := cis_eks.data_adapter.process_args

rule_evaluation {
	common.contains_key_with_value(process_args, "--feature-gates", "RotateKubeletServerCertificate=true")
}

# In case both flags and configuration file are specified, the executable argument takes precedence.
rule_evaluation {
	not process_args["--feature-gates"]
	data_adapter.process_config.config.featureGates.RotateKubeletServerCertificate
}

rule_evaluation {
	not contains(process_args["--feature-gates"], "RotateKubeletServerCertificate")
	data_adapter.process_config.config.featureGates.RotateKubeletServerCertificate
}

rule_evaluation {
	common.contains_key_with_value(process_args, "--rotate-server-certificates", "true")
}

rule_evaluation {
	data_adapter.process_config.config.serverTLSBootstrap
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
