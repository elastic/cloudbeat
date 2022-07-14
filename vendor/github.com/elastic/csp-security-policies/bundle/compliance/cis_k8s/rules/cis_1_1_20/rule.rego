package compliance.cis_k8s.rules.cis_1_1_20

import data.compliance.lib.common
import data.compliance.lib.data_adapter

# Ensure that the Kubernetes PKI certificate file permissions are set to 644 or more restrictive (Manual)
finding = result {
	# filter
	common.file_in_path("/etc/kubernetes/pki", data_adapter.file_path)
	endswith(data_adapter.filename, ".crt")

	# evaluate
	filemode := data_adapter.filemode
	rule_evaluation := common.file_permission_match(filemode, 6, 4, 4)

	# set result
	result := {
		"evaluation": common.calculate_result(rule_evaluation),
		"expected": {"filemode": "644"},
		"evidence": {"filemode": filemode},
	}
}
