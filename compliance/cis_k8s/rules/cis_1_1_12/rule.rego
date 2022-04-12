package compliance.cis_k8s.rules.cis_1_1_12

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the etcd data directory ownership is set to etcd:etcd
finding = result {
	common.file_in_path("/var/lib/etcd/", data_adapter.file_path)
	uid = data_adapter.owner_user_id
	gid = data_adapter.owner_group_id
	rule_evaluation := common.file_ownership_match(uid, gid, "etcd", "etcd")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"expected": {"uid": "etcd", "gid": "etcd"},
		"evidence": {"uid": uid, "gid": gid},
	}
}

metadata = {
	"name": "Ensure that the etcd data directory ownership is set to etcd:etcd",
	"description": "etcd is a highly-available key-value store used by Kubernetes deployments for persistent storage of all of its REST API objects. This data directory should be protected from any unauthorized reads or writes. It should be owned by etcd:etcd.",
	"impact": "None",
	"tags": array.concat(cis_k8s.default_tags, ["CIS 1.1.12", "Master Node Configuration"]),
	"benchmark": cis_k8s.benchmark_metadata,
	"remediation": "chown etcd:etcd /var/lib/etcd",
}
