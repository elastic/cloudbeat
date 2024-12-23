package compliance.cis_aws.rules.cis_4_6

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
			"FilterPattern": "{ ($.eventName = ConsoleLogin) && ($.errorMessage = \"Failed authentication\") }",
			"ParsedFilterPattern": pattern.complex_expression("&&", [
				pattern.simple_expression("$.errorMessage", "=", "\"Failed authentication\""),
				pattern.simple_expression("$.eventName", "=", "ConsoleLogin"),
			]),
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

rule_input(entry) := test_data.generate_monitoring_resources(entry)

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
