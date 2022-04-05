package compliance.cis_k8s.rules.cis_1_2_15

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the admission control plugin PodSecurityPolicy is set (Automated)
finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# evaluate
	process_args := cis_k8s.data_adapter.process_args
	rule_evaluation := common.arg_values_contains(process_args, "--enable-admission-plugins", "PodSecurityPolicy")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the admission control plugin PodSecurityPolicy is set",
	"description": "A Pod Security Policy is a cluster-level resource that controls the actions that a pod can perform and what it has the ability to access. The PodSecurityPolicy objects define a set of conditions that a pod must run with in order to be accepted into the system. Pod Security Policies are comprised of settings and strategies that control the security features a pod has access to and hence this must be used to control pod access permissions.",
	"impact": "The policy objects must be created and granted before pod creation would be allowed.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.15", "API Server"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the --disable-admission-plugins parameter to ensure it does not include NamespaceLifecycle.",
}
