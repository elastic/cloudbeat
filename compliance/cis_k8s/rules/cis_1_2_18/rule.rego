package compliance.cis_k8s.rules.cis_1_2_18

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --insecure-bind-address argument is not set (Automated)
finding = result {
	command_args := data_adapter.api_server_command_args
	rule_evaluation := common.contains_key(command_args, "--insecure-bind-address") == false

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the --insecure-bind-address argument is not set",
	"description": "The apiserver, by default, does not authenticate itself to the kubelet's HTTPS endpoints. The requests from the apiserver are treated anonymously. You should set up certificate-based kubelet authentication to ensure that the apiserver authenticates itself to kubelets when submitting requests.",
	"impact": "Connections to the API server will require valid authentication credentials.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.18", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Follow the Kubernetes documentation and set up the TLS connection between the apiserver and kubelets. Then, edit API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the kubelet client certificate and key parameters",
}
