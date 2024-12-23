package compliance.cis_gcp.rules.cis_6_2_8

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.sql.ensure_db_flag as audit
import future.keywords.if

# Ensure That 'cloudsql.enable_pgaudit' Database Flag for each Cloud Sql Postgresql Instance Is Set to 'on' For Centralized Logging.
finding := result if {
	# filter
	data_adapter.is_cloud_sql
	data_adapter.is_postgres_sql

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_flag_as_expected),
		{"DB Instance": data_adapter.resource},
	)
}

is_flag_as_expected := audit.is_flag_configured_as_expected("cloudsql.enable_pgaudit", ["on"])
