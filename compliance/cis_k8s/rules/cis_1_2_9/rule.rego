package compliance.cis_k8s.rules.cis_1_2_9

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --authorization-mode argument includes RBAC (Automated)
finding = result {
	# filter
	data_adapter.is_kube_apiserver

	# evaluate
	process_args := data_adapter.process_args
	rule_evaluation = common.arg_values_contains(process_args, "--authorization-mode", "RBAC")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"process_args": process_args},
	}
}

metadata = {
	"name": "Ensure that the --authorization-mode argument includes RBAC",
	"description": "Role Based Access Control (RBAC) allows fine-grained control over the operations that different entities can perform on different objects in the cluster. It is recommended to use the RBAC authorization mode.",
	"impact": "When RBAC is enabled you will need to ensure that appropriate RBAC settings (including Roles, RoleBindings and ClusterRoleBindings) are configured to allow appropriate access.",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.9", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the --authorization-mode parameter to a value that includes RBAC, for example: --authorization-mode=Node,RBAC",
}
