package compliance.cis_k8s.rules.cis_1_1_11

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the etcd data directory permissions are set to 700 or more restrictive
finding = result {
	common.file_in_path("/var/lib/etcd/", data_adapter.file_path)
	filemode := data_adapter.filemode
	rule_evaluation := common.file_permission_match(filemode, 7, 0, 0)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"filemode": filemode},
	}
}

metadata = {
	"name": "Ensure that the etcd data directory permissions are set to 700 or more restrictive",
	"description": "etcd is a highly-available key-value store used by Kubernetes deployments for persistent storage of all of its REST API objects. This data directory should be protected from any unauthorized reads or writes. It should not be readable or writable by any group members or the world.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.1.11", "Master Node Configuration"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "chmod 700 /var/lib/etcd",
}
