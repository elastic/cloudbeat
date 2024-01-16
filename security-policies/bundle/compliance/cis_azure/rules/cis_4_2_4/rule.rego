package compliance.cis_azure.rules.cis_4_2_4

import data.compliance.lib.common
import data.compliance.policy.azure.data_adapter
import data.compliance.policy.azure.sql_server.ensure_vulnerability_assessment_storage_account as audit
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
		audit.ensure_vulnerability_assessment_storage_account(setting)
		count(setting.notificationEmails) > 0
	}
}
