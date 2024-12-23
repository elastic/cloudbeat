package compliance.cis_aws.rules.cis_4_12

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
			"FilterPattern": "{ ($.eventName = CreateCustomerGateway) || ($.eventName = DeleteCustomerGateway) || ($.eventName = AttachInternetGateway) || ($.eventName = CreateInternetGateway) || ($.eventName = DeleteInternetGateway) || ($.eventName = DetachInternetGateway) }",
			"ParsedFilterPattern": pattern.complex_expression("||", [
				pattern.simple_expression("$.eventName", "=", "AttachInternetGateway"),
				pattern.simple_expression("$.eventName", "=", "CreateInternetGateway"),
				pattern.simple_expression("$.eventName", "=", "CreateCustomerGateway"),
				pattern.simple_expression("$.eventName", "=", "DeleteInternetGateway"),
				pattern.simple_expression("$.eventName", "=", "DeleteCustomerGateway"),
				pattern.simple_expression("$.eventName", "=", "DetachInternetGateway"),
			]),
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

rule_input(entry) := test_data.generate_monitoring_resources(entry)

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
