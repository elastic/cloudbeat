package compliance.cis_k8s.rules.cis_1_1_13

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the admin.conf file permissions are set to 600 or more restrictive
finding = result {
	data_adapter.filename == "admin.conf"
	filemode := data_adapter.filemode
	rule_evaluation := common.file_permission_match(filemode, 6, 0, 0)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"expected": {"filemode": "600"},
		"evidence": {"filemode": filemode},
	}
}
