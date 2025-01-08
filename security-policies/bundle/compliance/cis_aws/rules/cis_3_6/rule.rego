package compliance.cis_aws.rules.cis_3_6

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import future.keywords.if

default rule_evaluation := false

rule_evaluation if {
	data_adapter.trail_bucket_info.logging.Enabled == true
}

# Ensure S3 bucket access logging is enabled on the CloudTrail S3 bucket
finding := result if {
	# filter
	data_adapter.is_single_trail

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}
