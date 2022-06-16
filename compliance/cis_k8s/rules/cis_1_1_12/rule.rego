package compliance.cis_k8s.rules.cis_1_1_12

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the etcd data directory ownership is set to etcd:etcd
finding = result {
	common.file_in_path("/var/lib/etcd/", data_adapter.file_path)
	user = data_adapter.owner_user
	group = data_adapter.owner_group
	rule_evaluation := common.file_ownership_match(user, group, "etcd", "etcd")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"expected": {"owner": "etcd", "group": "etcd"},
		"evidence": {"owner": user, "group": group},
	}
}
