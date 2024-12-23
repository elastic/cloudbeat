package compliance.cis_aws.rules.cis_3_3

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import data.compliance.policy.aws_cloudtrail.no_public_bucket_access as audit
import future.keywords.if

# Ensure the S3 bucket used to store CloudTrail logs is not publicly accessible.
finding := result if {
	# filter
	data_adapter.is_single_trail

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(audit.bucket_is_public == false),
		input.resource,
	)
}
