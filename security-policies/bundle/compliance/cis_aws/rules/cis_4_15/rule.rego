package compliance.cis_aws.rules.cis_4_15

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

# { ($.eventSource = organizations.amazonaws.com) && (($.eventName = \"AcceptHandshake\") || ($.eventName = \"AttachPolicy\") || ($.eventName = \"CreateAccount\") || ($.eventName = \"CreateOrganizationalUnit\") || ($.eventName = \"CreatePolicy\") || ($.eventName = \"DeclineHandshake\") || ($.eventName = \"DeleteOrganization\") || ($.eventName = \"DeleteOrganizationalUnit\") || ($.eventName = \"DeletePolicy\") || ($.eventName = \"DetachPolicy\") || ($.eventName = \"DisablePolicyType\") || ($.eventName = \"EnablePolicyType\") || ($.eventName = \"InviteAccountToOrganization\") || ($.eventName = \"LeaveOrganization\") || ($.eventName = \"MoveAccount\") || ($.eventName = \"RemoveAccountFromOrganization\") || ($.eventName = \"UpdatePolicy\") || ($.eventName = \"UpdateOrganizationalUnit\")) }
required_patterns = [pattern.complex_expression("&&", [
	pattern.simple_expression("$.eventSource", "=", "organizations.amazonaws.com"),
	pattern.complex_expression("||", [
		pattern.simple_expression("$.eventName", "=", "\"AcceptHandshake\""),
		pattern.simple_expression("$.eventName", "=", "\"AttachPolicy\""),
		pattern.simple_expression("$.eventName", "=", "\"CreateAccount\""),
		pattern.simple_expression("$.eventName", "=", "\"CreateOrganizationalUnit\""),
		pattern.simple_expression("$.eventName", "=", "\"CreatePolicy\""),
		pattern.simple_expression("$.eventName", "=", "\"DeclineHandshake\""),
		pattern.simple_expression("$.eventName", "=", "\"DeleteOrganization\""),
		pattern.simple_expression("$.eventName", "=", "\"DeleteOrganizationalUnit\""),
		pattern.simple_expression("$.eventName", "=", "\"DeletePolicy\""),
		pattern.simple_expression("$.eventName", "=", "\"DetachPolicy\""),
		pattern.simple_expression("$.eventName", "=", "\"DisablePolicyType\""),
		pattern.simple_expression("$.eventName", "=", "\"EnablePolicyType\""),
		pattern.simple_expression("$.eventName", "=", "\"InviteAccountToOrganization\""),
		pattern.simple_expression("$.eventName", "=", "\"LeaveOrganization\""),
		pattern.simple_expression("$.eventName", "=", "\"MoveAccount\""),
		pattern.simple_expression("$.eventName", "=", "\"RemoveAccountFromOrganization\""),
		pattern.simple_expression("$.eventName", "=", "\"UpdatePolicy\""),
		pattern.simple_expression("$.eventName", "=", "\"UpdateOrganizationalUnit\""),
	]),
])]

rule_evaluation = trail.at_least_one_trail_satisfied(required_patterns)
