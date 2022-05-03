package compliance.cis_eks.rules.cis_3_1_2

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the kubelet kubeconfig file ownership is set to root:root
finding = result {
	# filter
	data_adapter.filename == "kubeconfig"

	# evaluate
	uid = data_adapter.owner_user_id
	gid = data_adapter.owner_group_id
	rule_evaluation := common.file_ownership_match(uid, gid, "root", "root")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"uid": uid, "gid": gid},
	}
}
