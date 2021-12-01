package compliance.cis_k8s.rules.cis_1_2_27

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --service-account-lookup argument is set to true (Automated)

# Verify that if the --service-account-lookup argument exists it is set to true.
command_args := data_adapter.api_server_command_args

default rule_evaluation = false

rule_evaluation {
	common.contains_key_with_value(command_args, "--service-account-lookup", "true")
} else {
	common.contains_key(command_args, "--service-account-lookup") == false
}

finding = result {
	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the --service-account-lookup argument is set to true",
	"description": "If --service-account-lookup is not enabled, the apiserver only verifies that the authentication token is valid, and does not validate that the service account token mentioned in the request is actually present in etcd. This allows using a service account token even after the corresponding service account is deleted. This is an example of time of check to time of use security issue.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.27", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the below parameter. --service-account-lookup=true",
}
