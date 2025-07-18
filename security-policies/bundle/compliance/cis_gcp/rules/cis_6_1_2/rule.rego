package compliance.cis_gcp.rules.cis_6_1_2

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

finding := result if {
	data_adapter.is_cloud_sql
	data_adapter.is_cloud_my_sql

	result := common.generate_evaluation_result(common.calculate_result(skip_show_database_enabled))
}

skip_show_database_enabled if {
	flags := data_adapter.resource.data.settings.databaseFlags[_]
	flags.name == "skip_show_database"
	flags.value == "on"
} else := false
