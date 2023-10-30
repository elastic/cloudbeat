package compliance.cis_azure.rules.cis_3_2

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.storage_account.ensure_encryption as audit

finding = result {
	# filter
	data_adapter.is_storage_account

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.is_encryption_enabled),
		{"Resource": data_adapter.resource},
	)
}
