package compliance.cis_k8s.rules.cis_1_1_19

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the Kubernetes PKI directory and file ownership is set to root:root
finding = result {
	common.file_in_path("/etc/kubernetes/pki/", data_adapter.file_path)

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
