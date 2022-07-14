package compliance.cis_eks.rules.cis_3_1_4

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the kubelet configuration file ownership is set to root:root
finding = result {
	# filter
	data_adapter.filename == "kubelet-config.json"

	# evaluate
	user = data_adapter.owner_user
	group = data_adapter.owner_group
	rule_evaluation := common.file_ownership_match(user, group, "root", "root")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"owner": user, "group": group},
	}
}
