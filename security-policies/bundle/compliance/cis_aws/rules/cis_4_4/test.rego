package compliance.cis_aws.rules.cis_4_4

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern
import data.lib.test
import future.keywords.if

test_violation if {
	# No items
	eval_fail with input as rule_input([])

	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{
			"FilterPattern": "{($.eventName=DeleteGroupPolicy)||($.eventName=DeleteRolePolicy)||($.eventName=DeleteUserPolicy)||($.eventName=PutGroupPolicy)||($.eventName=PutRolePolicy)||($.eventName=PutUserPolicy)||($.eventName=CreatePolicy)||($.eventName=DeletePolicy)||($.event Name=CreatePolicyVersion)||($.eventName=DeletePolicyVersion)||($.eventName=AttachRolePolicy)||($.eventName=DetachRolePolicy)||($.eventName=AttachUserPolicy)||($.eventName=DetachUserPolicy)||($.eventName=AttachGroupPolicy)}",
			"ParsedFilterPattern": pattern.complex_expression("||", [
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
				# One is missing
			]),
		}],
		"Topics": ["arn:aws:...sns"],
	}])
}

test_pass if {
	eval_pass with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{($.eventName=DeleteGroupPolicy)||($.eventName=DeleteRolePolicy)||($.eventName=DeleteUserPolicy)||($.eventName=PutGroupPolicy)||($.eventName=PutRolePolicy)||($.eventName=PutUserPolicy)||($.eventName=CreatePolicy)||($.eventName=DeletePolicy)||($.eventName=CreatePolicyVersion)||($.eventName=DeletePolicyVersion)||($.eventName=AttachRolePolicy)||($.eventName=DetachRolePolicy)||($.eventName=AttachUserPolicy)||($.eventName=DetachUserPolicy)||($.eventName=AttachGroupPolicy)||($.eventName=DetachGroupPolicy)}",
			"ParsedFilterPattern": pattern.complex_expression("||", [
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
			]),
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

rule_input(entry) = test_data.generate_monitoring_resources(entry)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
