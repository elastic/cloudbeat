package compliance.cis_gcp.rules.cis_6_7

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

finding := result if {
	data_adapter.is_sql_instance

	result := common.generate_result_without_expected(
		common.calculate_result(backup_enabled),
		data_adapter.resource,
	)
}

backup_enabled if {
	data_adapter.resource.data.settings.backupConfiguration.enabled == true
} else := false
