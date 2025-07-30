package compliance.cis_gcp.rules.cis_6_2_5

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.sql.ensure_db_flag as audit
import future.keywords.if

default is_flag_as_expected := false

# Ensure that the ‘Log_min_messages’ Flag for a Cloud SQL PostgreSQL Instance is set at minimum to 'Warning'.
finding := result if {
	# filter
	data_adapter.is_cloud_sql
	data_adapter.is_postgres_sql

	# set result
	result := common.generate_evaluation_result(common.calculate_result(is_flag_as_expected))
}

is_flag_as_expected if {
	# We followed the CIS benchmark recommendation and configured the expected value as "error"
	# although the recommendation could also be interpreted as permitting a range of rules.
	audit.is_flag_configured_as_expected("log_min_messages", ["error"])
}

is_flag_as_expected if {
	not audit.is_flag_exists("log_min_messages")
}
