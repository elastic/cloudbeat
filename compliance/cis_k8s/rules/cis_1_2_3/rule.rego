package compliance.cis_k8s.rules.cis_1_2_3

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --token-auth-file parameter is not set (Automated)
finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# evaluate
	process_args := data_adapter.process_args
	rule_evaluation := assert.is_false(common.contains_key(process_args, "--token-auth-file"))

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --token-auth-file parameter is not set",
	"description": "The token-based authentication utilizes static tokens to authenticate requests to the apiserver. The tokens are stored in clear-text in a file on the apiserver, and cannot be revoked or rotated without restarting the apiserver. Hence, do not use static token-based authentication.",
	"impact": "You will have to configure and use alternate authentication mechanisms such as tokens and certificates. Username and password for basic authentication could no longer be used.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.3", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Follow the documentation and configure alternate mechanisms for authentication. Then, edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and remove the --token-auth-file=<filename> parameter.",
}
