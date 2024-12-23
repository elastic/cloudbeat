package compliance.cis_aws.rules.cis_4_7

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
			"FilterPattern": "{($.eventSource = kms.amazonaws.com) && (($.eventName=DisableKey)||($.eventName=ScheduleKeyDeletion)) }",
			"ParsedFilterPattern": pattern.complex_expression("&&", [
				pattern.complex_expression("||", [
					pattern.simple_expression("$.eventName", "=", "ScheduleKeyDeletion"),
					pattern.simple_expression("$.eventName", "=", "DisableKey"),
				]),
				pattern.simple_expression("$.eventSource", "=", "kms.amazonaws.com"),
			]),
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

rule_input(entry) := test_data.generate_monitoring_resources(entry)

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
