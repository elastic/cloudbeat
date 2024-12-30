package compliance.cis_aws.rules.cis_3_10

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import data.compliance.policy.aws_cloudtrail.verify_s3_object_logging as audit
import future.keywords.if

# Ensure that Object-level logging for write events is enabled for S3 bucket.
finding := result if {
	# filter
	data_adapter.is_single_trail

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

rule_evaluation := audit.ensure_s3_object_logging(["All", "WriteOnly"])
