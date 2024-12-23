package compliance.cis_gcp.rules.cis_6_2_9

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.sql.ensure_private_ip as audit
import future.keywords.if

# Ensure Instance IP assignment is set to private.
finding := result if {
	# filter
	data_adapter.is_cloud_sql
	data_adapter.is_postgres_sql

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.ip_is_private),
		data_adapter.resource,
	)
}
