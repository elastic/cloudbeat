package compliance.cis_gcp.rules.cis_6_2_1

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.sql.ensure_db_flag as audit

default is_flag_as_expected = false

# Ensure ‘Log_error_verbosity’ Database Flag for Cloud SQL PostgreSQL Instance Is Set to ‘DEFAULT’ or Stricter
finding = result {
	# filter
	data_adapter.is_cloud_sql
	data_adapter.is_postgres_sql

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(is_flag_as_expected),
		{"DB Instance": data_adapter.resource},
	)
}

is_flag_as_expected {
	# We followed the CIS benchmark recommendation and configured the expected value as "default"
	# although the recommendation could also be interpreted as permitting a range of rules.
	audit.is_flag_configured_as_expected("log_error_verbosity", ["default"])
}

is_flag_as_expected {
	not audit.is_flag_exists("log_error_verbosity")
}
