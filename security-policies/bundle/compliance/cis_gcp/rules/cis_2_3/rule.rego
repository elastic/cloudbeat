package compliance.cis_gcp.rules.cis_2_3

import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import future.keywords.if

default is_retention_policy_valid := false

# Ensure That Retention Policies on Cloud Storage Buckets Used for Exporting Logs Are Configured Using Bucket Lock.
finding := result if {
	data_adapter.is_log_bucket

	result := common.generate_evaluation_result(common.calculate_result(is_retention_policy_valid))
}

is_retention_policy_valid if {
	data_adapter.resource.data.retentionDays > 0
	data_adapter.resource.data.locked == true
}
