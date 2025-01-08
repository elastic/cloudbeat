package compliance.cis_gcp.rules.cis_5_2

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

default rule_evaluation := false

# Ensure That Cloud Storage Buckets Have Uniform Bucket- Level Access Enabled.
finding := result if {
	# filter
	data_adapter.is_storage_bucket

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		{"GCS Bucket": input.resource},
	)
}

rule_evaluation if {
	not is_null(data_adapter.resource.data.iamConfiguration.uniformBucketLevelAccess.enabled)
	data_adapter.resource.data.iamConfiguration.uniformBucketLevelAccess.enabled
}
