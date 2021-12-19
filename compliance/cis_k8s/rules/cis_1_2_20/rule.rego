package compliance.cis_k8s.rules.cis_1_2_20

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --secure-port argument is not set to 0 (Automated)
finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# evaluate
	process_args := data_adapter.process_args
	rule_evaluation = assert.is_false(common.contains_key_with_value(process_args, "--secure-port", "0"))

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --secure-port argument is not set to 0",
	"description": "The secure port is used to serve https with authentication and authorization. If you disable it, no https traffic is served and all traffic is served unencrypted.",
	"impact": "You need to set the API Server up with the right TLS certificates.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.20", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and either remove the --secure-port parameter or set it to a different (non-zero) desired port.",
}
