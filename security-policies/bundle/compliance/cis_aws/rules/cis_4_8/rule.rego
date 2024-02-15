package compliance.cis_aws.rules.cis_4_8

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern
import data.compliance.policy.aws_cloudtrail.trail
import future.keywords.if

default rule_evaluation = false

finding = result if {
	# filter
	data_adapter.is_multi_trails_type

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

# { ($.eventSource = s3.amazonaws.com) && (($.eventName = PutBucketAcl) || ($.eventName = PutBucketPolicy) || ($.eventName = PutBucketCors) || ($.eventName = PutBucketLifecycle) || ($.eventName = PutBucketReplication) || ($.eventName = DeleteBucketPolicy) || ($.eventName = DeleteBucketCors) || ($.eventName = DeleteBucketLifecycle) || ($.eventName = DeleteBucketReplication)) }
required_patterns = [pattern.complex_expression("&&", [
	pattern.simple_expression("$.eventSource", "=", "s3.amazonaws.com"),
	pattern.complex_expression("||", [
		pattern.simple_expression("$.eventName", "=", "PutBucketAcl"),
		pattern.simple_expression("$.eventName", "=", "PutBucketPolicy"),
		pattern.simple_expression("$.eventName", "=", "PutBucketCors"),
		pattern.simple_expression("$.eventName", "=", "PutBucketLifecycle"),
		pattern.simple_expression("$.eventName", "=", "PutBucketReplication"),
		pattern.simple_expression("$.eventName", "=", "DeleteBucketPolicy"),
		pattern.simple_expression("$.eventName", "=", "DeleteBucketCors"),
		pattern.simple_expression("$.eventName", "=", "DeleteBucketLifecycle"),
		pattern.simple_expression("$.eventName", "=", "DeleteBucketReplication"),
	]),
])]

rule_evaluation = trail.at_least_one_trail_satisfied(required_patterns)
