package compliance.cis_gcp.rules.cis_7_1

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.iam.ensure_no_public_access as audit
import future.keywords.if

# Ensure That BigQuery Datasets Are Not Anonymously or Publicly Accessible.
finding := result if {
	# filter
	data_adapter.is_bigquery_dataset

	# set result
	result := common.generate_evaluation_result(common.calculate_result(assert.is_false(audit.resource_is_public)))
}
