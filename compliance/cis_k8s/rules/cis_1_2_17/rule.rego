package compliance.cis_k8s.rules.cis_1_2_17

import data.compliance.cis_k8s
import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --insecure-bind-address argument is not set (Automated)
finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# evaluate
	process_args := cis_k8s.data_adapter.process_args
	rule_evaluation := assert.is_false(common.contains_key(process_args, "--insecure-bind-address"))

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --insecure-bind-address argument is not set",
	"description": "The apiserver, by default, does not authenticate itself to the kubelet's HTTPS endpoints. The requests from the apiserver are treated anonymously. You should set up certificate-based kubelet authentication to ensure that the apiserver authenticates itself to kubelets when submitting requests.",
	"impact": "Connections to the API server will require valid authentication credentials.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.17", "API Server"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "Follow the Kubernetes documentation and set up the TLS connection between the apiserver and kubelets. Then, edit API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the kubelet client certificate and key parameters",
}
