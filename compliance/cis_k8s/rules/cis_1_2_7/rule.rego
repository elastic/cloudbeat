package compliance.cis_k8s.rules.cis_1_2_7

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --authorization-mode argument is not set to AlwaysAllow (Automated)
finding = result {
	command_args := data_adapter.api_server_command_args
	rule_evaluation = common.arg_values_contains(command_args, "--authorization-mode", "AlwaysAllow") == false

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the --authorization-mode argument is not set to AlwaysAllow",
	"description": "The API Server, can be configured to allow all requests. This mode should not be used on any production cluster.",
	"impact": "Only authorized requests will be served.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.7", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the --authorization-mode parameter to values other than AlwaysAllow",
}
