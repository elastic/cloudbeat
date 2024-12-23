package compliance.cis_aws.rules.cis_4_10

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

# { ($.eventName = AuthorizeSecurityGroupIngress) || ($.eventName = AuthorizeSecurityGroupEgress) || ($.eventName = RevokeSecurityGroupIngress) || ($.eventName = RevokeSecurityGroupEgress) || ($.eventName = CreateSecurityGroup) || ($.eventName = DeleteSecurityGroup) }
required_patterns := [pattern.complex_expression("||", [
	pattern.simple_expression("$.eventName", "=", "AuthorizeSecurityGroupIngress"),
	pattern.simple_expression("$.eventName", "=", "AuthorizeSecurityGroupEgress"),
	pattern.simple_expression("$.eventName", "=", "RevokeSecurityGroupIngress"),
	pattern.simple_expression("$.eventName", "=", "RevokeSecurityGroupEgress"),
	pattern.simple_expression("$.eventName", "=", "CreateSecurityGroup"),
	pattern.simple_expression("$.eventName", "=", "DeleteSecurityGroup"),
])]

rule_evaluation := trail.at_least_one_trail_satisfied(required_patterns)
