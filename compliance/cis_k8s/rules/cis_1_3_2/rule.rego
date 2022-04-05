package compliance.cis_k8s.rules.cis_1_3_2

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --profiling argument is set to false (Automated)
finding = result {
	# filter
	data_adapter.is_kube_controller_manger

	# evaluate
	process_args := cis_k8s.data_adapter.process_args
	rule_evaluation := common.contains_key_with_value(process_args, "--profiling", "false")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --profiling argument is set to false",
	"description": "Profiling allows for the identification of specific performance bottlenecks. It generates a significant amount of program data that could potentially be exploited to uncover system and program details. If you are not experiencing any bottlenecks and do not need the profiler for troubleshooting purposes, it is recommended to turn it off to reduce the potential attack surface.",
	"impact": "Profiling information would not be available.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.3.2", "Controller Manager"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "Edit the Controller Manager pod specification file /etc/kubernetes/manifests/kube-controller-manager.yaml on the master node and set to --profiling=false",
}
