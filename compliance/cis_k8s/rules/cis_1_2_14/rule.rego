package compliance.cis_k8s.rules.cis_1_2_14

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the admission control plugin ServiceAccount is set (Automated)
command_args := data_adapter.api_server_command_args

default rule_evaluation = false

rule_evaluation {
	# Verify that the --disable-admission-plugins argument is set to a value that does not includes ServiceAccount.
	command_args["--disable-admission-plugins"]
	not common.arg_values_contains(command_args, "--disable-admission-plugins", "ServiceAccount")
}

finding = result {
	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the admission control plugin ServiceAccount is set",
	"description": "When you create a pod, if you do not specify a service account, it is automatically assigned the default service account in the same namespace. You should create your own service account and let the API server manage its security tokens.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.14", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Follow the documentation and create ServiceAccount objects as per your environment. Then, edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and ensure that the --disable-admission-plugins parameter is set to a value that does not include ServiceAccount.",
}
