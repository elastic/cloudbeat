package compliance.cis_k8s.rules.cis_1_2_25

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate (Automated)
command_args := data_adapter.api_server_command_args

default rule_evaluation = false

rule_evaluation {
	value := command_args["--audit-log-maxsize"]
	common.greater_or_equal(value, 100)
}

finding = result {
	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"command_args": command_args},
	}
}

metadata = {
	"name": "Ensure that the --audit-log-maxsize argument is set to 100 or as appropriate",
	"description": "Kubernetes automatically rotates the log files. Retaining old log files ensures that you would have sufficient log data available for carrying out any investigation or correlation. If you have set file size of 100 MB and the number of old log files to keep as 10, you would approximate have 1 GB of log data that you could potentially use for your analysis.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.25", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the --audit-log-maxsize parameter to an appropriate size in MB. For example, to set it as 100 MB: --audit-log-maxsize=100",
}
