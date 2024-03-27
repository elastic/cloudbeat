package compliance.cis_aws.rules.cis_4_15

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern
import data.lib.test
import future.keywords.if

test_pass if {
	eval_pass with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.eventSource = organizations.amazonaws.com) && (($.eventName = \"AcceptHandshake\") || ($.eventName = \"AttachPolicy\") || ($.eventName = \"CreateAccount\") || ($.eventName = \"CreateOrganizationalUnit\") || ($.eventName = \"CreatePolicy\") || ($.eventName = \"DeclineHandshake\") || ($.eventName = \"DeleteOrganization\") || ($.eventName = \"DeleteOrganizationalUnit\") || ($.eventName = \"DeletePolicy\") || ($.eventName = \"DetachPolicy\") || ($.eventName = \"DisablePolicyType\") || ($.eventName = \"EnablePolicyType\") || ($.eventName = \"InviteAccountToOrganization\") || ($.eventName = \"LeaveOrganization\") || ($.eventName = \"MoveAccount\") || ($.eventName = \"RemoveAccountFromOrganization\") || ($.eventName = \"UpdatePolicy\") || ($.eventName = \"UpdateOrganizationalUnit\")) }",
			"ParsedFilterPattern": pattern.complex_expression("&&", [
				pattern.complex_expression("||", [
					pattern.simple_expression("$.eventName", "=", "\"CreateAccount\""),
					pattern.simple_expression("$.eventName", "=", "\"DeleteOrganizationalUnit\""),
					pattern.simple_expression("$.eventName", "=", "\"CreateOrganizationalUnit\""),
					pattern.simple_expression("$.eventName", "=", "\"CreatePolicy\""),
					pattern.simple_expression("$.eventName", "=", "\"DeclineHandshake\""),
					pattern.simple_expression("$.eventName", "=", "\"DeleteOrganization\""),
					pattern.simple_expression("$.eventName", "=", "\"AttachPolicy\""),
					pattern.simple_expression("$.eventName", "=", "\"DeletePolicy\""),
					pattern.simple_expression("$.eventName", "=", "\"DetachPolicy\""),
					pattern.simple_expression("$.eventName", "=", "\"DisablePolicyType\""),
					pattern.simple_expression("$.eventName", "=", "\"EnablePolicyType\""),
					pattern.simple_expression("$.eventName", "=", "\"AcceptHandshake\""),
					pattern.simple_expression("$.eventName", "=", "\"InviteAccountToOrganization\""),
					pattern.simple_expression("$.eventName", "=", "\"LeaveOrganization\""),
					pattern.simple_expression("$.eventName", "=", "\"MoveAccount\""),
					pattern.simple_expression("$.eventName", "=", "\"RemoveAccountFromOrganization\""),
					pattern.simple_expression("$.eventName", "=", "\"UpdatePolicy\""),
					pattern.simple_expression("$.eventName", "=", "\"UpdateOrganizationalUnit\""),
				]),
				pattern.simple_expression("$.eventSource", "=", "organizations.amazonaws.com"),
			]),
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

rule_input(entry) = test_data.generate_monitoring_resources(entry)

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
