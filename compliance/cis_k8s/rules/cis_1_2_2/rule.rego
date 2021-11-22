package compliance.cis_k8s.rules.cis_1_2_2

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --basic-auth-file argument is not set (Automated)
finding = result {
	command_args := data_adapter.command_args
	rule_evaluation := common.array_contains(command_args, "--basic-auth-file") == false

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the --basic-auth-file argument is not set",
	"description": "Basic authentication uses plaintext credentials for authentication. Currently, the basic authentication credentials last indefinitely, and the password cannot be changed without restarting the API server. The basic authentication is currently supported for convenience. Hence, basic authentication should not be used.",
	"impact": "You will have to configure and use alternate authentication mechanisms such as tokens and certificates. Username and password for basic authentication could no longer be used.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.2", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Follow the documentation and configure alternate mechanisms for authentication. Then, edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and remove the --basic-auth-file=<filename> parameter.",
}
