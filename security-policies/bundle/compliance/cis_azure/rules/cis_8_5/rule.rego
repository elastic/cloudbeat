package compliance.cis_azure.rules.cis_8_5

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_vault

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_vault_recoverable),
		{"Resource": data_adapter.resource},
	)
}

is_vault_recoverable if {
	data_adapter.properties.enableSoftDelete
	data_adapter.properties.enablePurgeProtection
} else := false
