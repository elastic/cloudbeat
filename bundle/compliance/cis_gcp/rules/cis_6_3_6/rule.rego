package compliance.cis_gcp.rules.cis_6_3_6

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.sql.ensure_db_flag as audit

finding = result {
	# filter
	data_adapter.is_cloud_sql
	data_adapter.is_sql_server

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_flag_enabled),
		{"DB Instance": data_adapter.resource},
	)
}

is_flag_enabled := audit.is_flag_enabled("3625") #trace flag
