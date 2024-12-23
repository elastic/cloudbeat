package compliance.cis_aws.rules.cis_4_12

import data.compliance.lib.common
import data.compliance.policy.aws_cloudtrail.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern
import data.compliance.policy.aws_cloudtrail.trail
import future.keywords.if

default rule_evaluation := false

finding := result if {
	# filter
	data_adapter.is_multi_trails_type

	# set result
	result := common.generate_result_without_expected(
		common.calculate_result(rule_evaluation),
		input.resource,
	)
}

# { ($.eventName = CreateCustomerGateway) || ($.eventName = DeleteCustomerGateway) || ($.eventName = AttachInternetGateway) || ($.eventName = CreateInternetGateway) || ($.eventName = DeleteInternetGateway) || ($.eventName = DetachInternetGateway) }
required_patterns := [pattern.complex_expression("||", [
	pattern.simple_expression("$.eventName", "=", "CreateCustomerGateway"),
	pattern.simple_expression("$.eventName", "=", "DeleteCustomerGateway"),
	pattern.simple_expression("$.eventName", "=", "AttachInternetGateway"),
	pattern.simple_expression("$.eventName", "=", "CreateInternetGateway"),
	pattern.simple_expression("$.eventName", "=", "DeleteInternetGateway"),
	pattern.simple_expression("$.eventName", "=", "DetachInternetGateway"),
])]

rule_evaluation := trail.at_least_one_trail_satisfied(required_patterns)
