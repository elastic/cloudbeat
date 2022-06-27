package compliance.cis_k8s.rules.cis_1_1_7

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
		"expected": {"filemode": "644"},
		"evidence": {"filemode": filemode},
	}
}
