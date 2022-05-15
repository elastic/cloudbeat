package compliance.cis_eks.rules.cis_3_1_3

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the kubelet configuration file has permissions set to 644 or more restrictive
finding = result {
	# filter
	data_adapter.filename == "kubelet-config.json"

	# evaluate
	filemode := data_adapter.filemode
	rule_evaluation := common.file_permission_match(filemode, 6, 4, 4)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"filemode": filemode},
	}
}
