package compliance.cis_k8s.rules.cis_1_2_15

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the admission control plugin NamespaceLifecycle is set (Automated)
command_args := data_adapter.api_server_command_args

default rule_evaluation = false

rule_evaluation {
	# Verify that the --disable-admission-plugins argument is set to a value that does not include NamespaceLifecycle.
	command_args["--disable-admission-plugins"]
	not common.arg_values_contains(command_args, "--disable-admission-plugins", "NamespaceLifecycle")
}

finding = result {
	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the admission control plugin NamespaceLifecycle is set",
	"description": "Setting admission control policy to NamespaceLifecycle ensures that objects cannot be created in non-existent namespaces, and that namespaces undergoing termination are not used for creating the new objects. This is recommended to enforce the integrity of the namespace termination process and also for the availability of the newer objects.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.15", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the --disable-admission-plugins parameter to ensure it does not include NamespaceLifecycle.",
}
