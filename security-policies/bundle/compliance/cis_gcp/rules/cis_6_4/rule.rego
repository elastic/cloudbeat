package compliance.cis_gcp.rules.cis_6_4

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if
import future.keywords.in

finding := result if {
	data_adapter.is_sql_instance
	is_relevant_sql_instance

	result := common.generate_evaluation_result(common.calculate_result(ssl_is_required))
}

ssl_is_required if {
	data_adapter.resource.data.settings.ipConfiguration.requireSsl == true
} else := false

is_relevant_sql_instance if {
	startswith(data_adapter.resource.data.databaseVersion, "POSTGRES")
} else if {
	startswith(data_adapter.resource.data.databaseVersion, "MYSQL")
} else if {
	startswith(data_adapter.resource.data.databaseVersion, "SQLSERVER_2017")
} else := false
