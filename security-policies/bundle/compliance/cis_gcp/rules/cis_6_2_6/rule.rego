package compliance.cis_gcp.rules.cis_6_2_6

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.sql.ensure_db_flag as audit
import future.keywords.if

default is_flag_as_expected := false

# Ensure ‘Log_min_error_statement’ Database Flag for Cloud SQL PostgreSQL Instance Is Set to ‘Error’ or Stricter.
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

is_flag_as_expected if {
	# We followed the CIS benchmark recommendation and configured the expected value as "error"
	# although the recommendation could also be interpreted as permitting a range of rules.
	audit.is_flag_configured_as_expected("log_min_error_statement", ["error"])
}

is_flag_as_expected if {
	not audit.is_flag_exists("log_min_error_statement")
}
