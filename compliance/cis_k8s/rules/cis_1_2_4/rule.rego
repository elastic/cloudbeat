package compliance.cis_k8s.rules.cis_1_2_4

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate (Automated)

# evaluate
process_args := cis_k8s.data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	process_args["--kubelet-client-certificate"]
	process_args["--kubelet-client-key"]
}

finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --kubelet-client-certificate and --kubelet-client-key arguments are set as appropriate",
	"description": "The apiserver, by default, does not authenticate itself to the kubelet's HTTPS endpoints. The requests from the apiserver are treated anonymously. You should set up certificate-based kubelet authentication to ensure that the apiserver authenticates itself to kubelets when submitting requests.",
	"impact": "You require TLS to be configured on apiserver as well as kubelets.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.4", "API Server"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "Follow the Kubernetes documentation and set up the TLS connection between the apiserver and kubelets. Then, edit API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the kubelet client certificate and key parameters",
}
