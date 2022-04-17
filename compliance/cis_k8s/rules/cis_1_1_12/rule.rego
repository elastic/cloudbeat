package compliance.cis_k8s.rules.cis_1_1_12

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
