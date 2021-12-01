package compliance.cis_k8s.rules.cis_1_2_22

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --audit-log-path argument is set (Automated)
finding = result {
	command_args := data_adapter.api_server_command_args
	rule_evaluation := common.contains_key(command_args, "--audit-log-path")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the --audit-log-path argument is set",
	"description": "Auditing the Kubernetes API Server provides a security-relevant chronological set of records documenting the sequence of activities that have affected system by individual users, administrators or other components of the system. Even though currently, Kubernetes provides only basic audit capabilities, it should be enabled. You can enable it by setting an appropriate audit log path.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.22", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the --audit-log-path parameter to a suitable path and file where you would like audit logs to be written, for example: --audit-log-path=/var/log/apiserver/audit.log",
}
