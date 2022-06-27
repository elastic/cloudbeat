package compliance.cis_k8s.rules.cis_1_1_4

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the controller manager pod specification file ownership is set to root:root (Automated)
finding = result {
	data_adapter.filename == "kube-controller-manager.yaml"
	user = data_adapter.owner_user
	group = data_adapter.owner_group
	rule_evaluation := common.file_ownership_match(user, group, "root", "root")

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"expected": {"owner": "root", "group": "root"},
		"evidence": {"owner": user, "group": group},
	}
}
