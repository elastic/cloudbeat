package compliance.cis_azure.rules.cis_7_3

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.disk.ensure_encryption as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_attached_disk

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.is_encryption_enabled),
		{"Resource": data_adapter.resource},
	)
}
