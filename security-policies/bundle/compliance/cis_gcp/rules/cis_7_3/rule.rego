package compliance.cis_gcp.rules.cis_7_3

import data.compliance.lib.common
import data.compliance.policy.gcp.bq.ensure_cmek_is_used as audit
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

# Ensure That a Default Customer-Managed Encryption Key (CMEK) Is Specified for All BigQuery Data Sets.
finding := result if {
	# filter
	data_adapter.is_bigquery_dataset

	# set result
	result := common.generate_evaluation_result(common.calculate_result(audit.is_cmek_used))
}
