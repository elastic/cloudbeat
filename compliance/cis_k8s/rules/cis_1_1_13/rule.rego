package compliance.cis_k8s.rules.cis_1_1_13

import data.compliance.cis_k8s
import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the admin.conf file permissions are set to 644 or more restrictive
finding = result {
	data_adapter.filename == "admin.conf"
	filemode := data_adapter.filemode
	rule_evaluation := common.file_permission_match(filemode, 6, 4, 4)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"evidence": {"filemode": filemode},
		"rule_name": "Ensure that the admin.conf file permissions are set to 644 or more restrictive",
		"tags": array.concat(cis_k8s.default_tags, ["CIS 1.1.13"]),
	}
}
