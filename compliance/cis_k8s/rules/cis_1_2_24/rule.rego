package compliance.cis_k8s.rules.cis_1_2_24

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate (Automated)

# evaluate
process_args := data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	value := process_args["--audit-log-maxbackup"]
	common.greater_or_equal(value, 10)
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
	"name": "Ensure that the --audit-log-maxbackup argument is set to 10 or as appropriate",
	"description": "Kubernetes automatically rotates the log files. Retaining old log files ensures that you would have sufficient log data available for carrying out any investigation or correlation. For example, if you have set file size of 100 MB and the number of old log files to keep as 10, you would approximate have 1 GB of log data that you could potentially use for your analysis.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.24", "API Server"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the --audit-log-maxbackup parameter to 10 or to an appropriate value --audit-log-maxbackup=10",
}
