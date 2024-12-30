package compliance.cis_azure.rules.cis_3_15

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.storage_account.ensure_tls_version as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_storage_account

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.is_tls_configured("TLS1_2")),
		{"Resource": data_adapter.resource},
	)
}
