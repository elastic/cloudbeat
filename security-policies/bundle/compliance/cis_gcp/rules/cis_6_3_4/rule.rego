package compliance.cis_gcp.rules.cis_6_3_4

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.sql.ensure_db_flag as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_cloud_sql
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(assert.is_false(is_flag_exists)),
		{"DB Instance": data_adapter.resource},
	)
}

is_flag_exists := audit.is_flag_exists("user options")
