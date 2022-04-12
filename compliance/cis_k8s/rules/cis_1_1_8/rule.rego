package compliance.cis_k8s.rules.cis_1_1_8

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the etcd pod specification file ownership is set to root:root (Automated)
finding = result {
	data_adapter.filename == "etcd.yaml"
	uid = data_adapter.owner_user_id
	gid = data_adapter.owner_group_id
	rule_evaluation := common.file_ownership_match(uid, gid, "root", "root")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"expected": {"uid": "root", "gid": "root"},
		"evidence": {"uid": uid, "gid": gid},
	}
}

metadata = {
	"name": "Ensure that the etcd pod specification file ownership is set to root:root",
	"description": "The etcd pod specification file /etc/kubernetes/manifests/etcd.yaml controls various parameters that set the behavior of the etcd service in the master node. etcd is a highly available key-value store which Kubernetes uses for persistent storage of all of its REST API object. You should set its file ownership to maintain the integrity of the file. The file should be owned by root:root.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.1.8", "Master Node Configuration"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "chown root:root /etc/kubernetes/manifests/etcd.yaml",
}
