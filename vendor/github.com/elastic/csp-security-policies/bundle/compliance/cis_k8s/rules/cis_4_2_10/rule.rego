package compliance.cis_k8s.rules.cis_4_2_10

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --tls-cert-file and --tls-private-key-file arguments are set as appropriate
process_args := cis_k8s.data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	process_args["--tls-cert-file"]
	process_args["--tls-private-key-file"]
}

rule_evaluation {
	data_adapter.process_config.config.tlsCertFile
	data_adapter.process_config.config.tlsPrivateKeyFile
}

finding = result {
	# filter
	data_adapter.is_kubelet

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}
