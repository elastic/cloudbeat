package compliance.cis_azure.rules.cis_5_2_9

import data.compliance.lib.common
import data.compliance.policy.azure.activity_log_alert.activity_log_alert_operation_enabled as audit
import data.compliance.policy.azure.data_adapter
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_activity_log_alerts

	operations = ["Microsoft.Network/publicIPAddresses/write"]
	categories = ["Administrative"]

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.activity_log_alert_operation_enabled(operations, categories)),
		{"Resource": data_adapter.resource},
	)
}
