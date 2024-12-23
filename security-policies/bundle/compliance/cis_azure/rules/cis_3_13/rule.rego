package compliance.cis_azure.rules.cis_3_13

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.storage_account.ensure_service_log as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_storage_account

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(logs_are_enabled),
		{"Resource": data_adapter.resource},
	)
}

default logs_are_enabled := false

logs_are_enabled if {
	audit.service_diagnostic_settings_log_rwd_enabled(data_adapter.resource.extension.blobDiagnosticSettings)
}
