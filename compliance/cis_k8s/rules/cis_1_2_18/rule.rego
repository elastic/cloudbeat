package compliance.cis_k8s.rules.cis_1_2_18

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --insecure-port argument is set to 0 (Automated)
finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# evaluate
	process_args := data_adapter.process_args
	rule_evaluation := common.contains_key_with_value(process_args, "--insecure-port", "0")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --insecure-port argument is set to 0",
	"description": "Setting up the apiserver to serve on an insecure port would allow unauthenticated and unencrypted access to your master node. This would allow attackers who could access this port, to easily take control of the cluster.",
	"impact": "All components that use the API must connect via the secured port, authenticate themselves, and be authorized to use the API. Including kube-controller-manage, kube-proxy, kube-scheduler, kubelets",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.18", "API Server"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set to --insecure-port=0.",
}
