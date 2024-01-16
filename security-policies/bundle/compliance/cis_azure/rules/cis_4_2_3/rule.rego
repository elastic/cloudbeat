package compliance.cis_azure.rules.cis_4_2_3

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import future.keywords.every
import future.keywords.if

finding = result if {
	# filter
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(va_recurrent_scans_enabled),
		{"Resource": data_adapter.resource},
	)
}

default va_recurrent_scans_enabled = false

va_recurrent_scans_enabled if {
	count(data_adapter.resource.extension.sqlVulnerabilityAssessmentSettings) > 0

	every setting in data_adapter.resource.extension.sqlVulnerabilityAssessmentSettings {
		not setting.storageAccountName == null
		not trim(setting.storageAccountName, " ") == ""
		setting.recurringScansEnabled == true
	}
}
