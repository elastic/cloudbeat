package compliance.cis_gcp.rules.cis_6_1_2

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter

finding = result {
	data_adapter.is_cloud_sql
	data_adapter.is_cloud_my_sql

	result := common.generate_result_without_expected(
		common.calculate_result(skip_show_database_enabled),
		data_adapter.resource,
	)
}

skip_show_database_enabled {
	flags := data_adapter.resource.data.settings.databaseFlags[_]
	flags.name == "skip_show_database"
	flags.value == "on"
} else = false
