package compliance.cis_gcp.rules.cis_5_1

import data.compliance.lib.assert
import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.compliance.policy.gcp.iam.ensure_no_public_access as audit
import future.keywords.if

# Ensure That Cloud Storage Bucket Is Not Anonymously or Publicly Accessible.
finding := result if {
	# filter
	data_adapter.is_storage_bucket

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(assert.is_false(audit.resource_is_public)),
		{"GCS Bucket": input.resource},
	)
}
