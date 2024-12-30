package compliance.cis_aws.rules.cis_3_11

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import data.compliance.policy.aws_cloudtrail.verify_s3_object_logging as audit
import future.keywords.if

# Ensure that Object-level logging for read events is enabled for S3 bucket.
finding := result if {
	# filter
	data_adapter.is_single_trail

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_eveluation),
		input.resource,
	)
}

rule_eveluation := audit.ensure_s3_object_logging(["All", "ReadOnly"])
