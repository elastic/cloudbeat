package compliance.cis_aws.rules.cis_4_4

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

# {($.eventName=DeleteGroupPolicy)||($.eventName=DeleteRolePolicy)||($.eventName=DeleteUserPolicy)||($.eventName=PutGroupPolicy)||($.eventName=PutRolePolicy)||($.eventName=PutUserPolicy)||($.eventName=CreatePolicy)||($.eventName=DeletePolicy)||($.eventName=CreatePolicyVersion)||($.eventName=DeletePolicyVersion)||($.eventName=AttachRolePolicy)||($.eventName=DetachRolePolicy)||($.eventName=AttachUserPolicy)||($.eventName=DetachUserPolicy)||($.eventName=AttachGroupPolicy)||($.eventName=DetachGroupPolicy)}
required_patterns = [pattern.complex_expression("||", [
	pattern.simple_expression("$.eventName", "=", "DeleteGroupPolicy"),
	pattern.simple_expression("$.eventName", "=", "DeleteRolePolicy"),
	pattern.simple_expression("$.eventName", "=", "DeleteUserPolicy"),
	pattern.simple_expression("$.eventName", "=", "PutGroupPolicy"),
	pattern.simple_expression("$.eventName", "=", "PutRolePolicy"),
	pattern.simple_expression("$.eventName", "=", "PutUserPolicy"),
	pattern.simple_expression("$.eventName", "=", "CreatePolicy"),
	pattern.simple_expression("$.eventName", "=", "DeletePolicy"),
	pattern.simple_expression("$.eventName", "=", "CreatePolicyVersion"),
	pattern.simple_expression("$.eventName", "=", "DeletePolicyVersion"),
	pattern.simple_expression("$.eventName", "=", "AttachRolePolicy"),
	pattern.simple_expression("$.eventName", "=", "DetachRolePolicy"),
	pattern.simple_expression("$.eventName", "=", "AttachUserPolicy"),
	pattern.simple_expression("$.eventName", "=", "DetachUserPolicy"),
	pattern.simple_expression("$.eventName", "=", "AttachGroupPolicy"),
	pattern.simple_expression("$.eventName", "=", "DetachGroupPolicy"),
])]

rule_evaluation = trail.at_least_one_trail_satisfied(required_patterns)
