package compliance.cis_k8s.rules.cis_1_1_7

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the etcd pod specification file permissions are set to 644 or more restrictive (Automated)
finding = result {
	data_adapter.filename == "etcd.yaml"
	filemode := data_adapter.filemode
	rule_evaluation := common.file_permission_match(filemode, 6, 4, 4)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"filemode": filemode},
	}
}

metadata = {
	"name": "Ensure that the etcd pod specification file permissions are set to 644 or more restrictive",
	"description": "The etcd pod specification file /etc/kubernetes/manifests/etcd.yaml controls various parameters that set the behavior of the etcd service in the master node. etcd is a highly available key-value store which Kubernetes uses for persistent storage of all of its REST API object. You should restrict its file permissions to maintain the integrity of the file. The file should be writable by only the administrators on the system.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.1.7", "Master Node Configuration"]),
	"benchmark": cis_k8s.benchmark_name,
	"remediation": "chmod 644 /etc/kubernetes/manifests/etcd.yaml",
}
