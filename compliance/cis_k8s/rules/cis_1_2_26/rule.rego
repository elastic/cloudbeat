package compliance.cis_k8s.rules.cis_1_2_26

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --etcd-certfile and --etcd-keyfile arguments are set as appropriate (Automated)

# evaluate
process_args := cis_k8s.data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	process_args["--etcd-certfile"]
	process_args["--etcd-keyfile"]
}

finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}
