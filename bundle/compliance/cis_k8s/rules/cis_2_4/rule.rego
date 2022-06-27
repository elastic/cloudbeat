package compliance.cis_k8s.rules.cis_2_4

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --peer-cert-file and --peer-key-file arguments are set as appropriate (Automated)

# evaluate
process_args := cis_k8s.data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	process_args["--peer-cert-file"]
	process_args["--peer-key-file"]
}

finding = result {
	# filter
	data_adapter.is_etcd

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}
