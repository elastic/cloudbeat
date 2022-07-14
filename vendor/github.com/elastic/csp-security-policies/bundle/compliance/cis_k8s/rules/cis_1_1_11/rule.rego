package compliance.cis_k8s.rules.cis_1_1_11

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
		"expected": {"filemode": "700"},
		"evidence": {"filemode": filemode},
	}
}
