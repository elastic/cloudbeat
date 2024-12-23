package compliance.cis_azure.rules.cis_3_8

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.storage_account.ensure_default_network_access as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_storage_account

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.is_default_network_access),
		{"Resource": data_adapter.resource},
	)
}
