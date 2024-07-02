package compliance.cis_aws.rules.cis_4_1

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

# { ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }
required_pattern = pattern.complex_expression("||", [
	pattern.simple_expression("$.errorCode", "=", "\"*UnauthorizedOperation\""),
	pattern.simple_expression("$.errorCode", "=", "\"AccessDenied*\""),
	pattern.simple_expression("$.sourceIPAddress", "!=", "\"delivery.logs.amazonaws.com\""),
	pattern.simple_expression("$.eventName", "!=", "\"HeadBucket\""),
])

rule_evaluation = trail.at_least_one_trail_satisfied([required_pattern])
