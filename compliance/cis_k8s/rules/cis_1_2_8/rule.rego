package compliance.cis_k8s.rules.cis_1_2_8

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --authorization-mode argument includes Node (Automated)
finding = result {
	command_args := data_adapter.api_server_command_args
	rule_evaluation = common.arg_values_contains(command_args, "--authorization-mode", "Node")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the --authorization-mode argument includes Node",
	"description": "The Node authorization mode only allows kubelets to read Secret, ConfigMap, PersistentVolume, and PersistentVolumeClaim objects associated with their nodes.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.8", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the --authorization-mode parameter to a value that includes Node.",
}
