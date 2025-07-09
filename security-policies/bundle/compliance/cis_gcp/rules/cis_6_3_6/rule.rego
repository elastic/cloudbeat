package compliance.cis_gcp.rules.cis_6_3_6

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.sql.ensure_db_flag as audit
import future.keywords.if

finding := result if {
	# filter
	data_adapter.is_cloud_sql
	data_adapter.is_sql_server

	# set result
	result := common.generate_evaluation_result(common.calculate_result(is_flag_configured_as_expected))
}

is_flag_configured_as_expected := audit.is_flag_configured_as_expected("3625", ["on"]) # trace flag
