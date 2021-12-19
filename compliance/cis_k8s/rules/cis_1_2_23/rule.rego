package compliance.cis_k8s.rules.cis_1_2_23

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the --audit-log-maxage argument is set to 30 or as appropriate (Automated)

# evaluate
process_args := data_adapter.process_args

default rule_evaluation = false

rule_evaluation {
	value := process_args["--audit-log-maxage"]
	common.greater_or_equal(value, 30)
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
	"name": "Ensure that the --audit-log-maxage argument is set to 30 or as appropriate",
	"description": "Retaining logs for at least 30 days ensures that you can go back in time and investigate or correlate any events. Set your audit log retention period to 30 days or as per your business requirements.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.2.23", "API Server"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "Edit the API server pod specification file /etc/kubernetes/manifests/kube-apiserver.yaml on the master node and set the --audit-log-maxage parameter to 30 or as an appropriate number of days: --audit-log-maxage=30",
}
