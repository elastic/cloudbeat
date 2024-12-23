package compliance.cis_azure.rules.cis_3_9

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.storage_account.ensure_service as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_storage_account

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.evaluate_service("AzureServices")),
		{"Resource": data_adapter.resource},
	)
}
