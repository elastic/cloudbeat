package compliance.cis_aws.rules.cis_4_3

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
			"FilterPattern": "{ $.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS && $.eventType != \"AwsServiceEvent\" }",
			"ParsedFilterPattern": pattern.complex_expression("&&", [
				pattern.simple_expression("$.eventType", "!=", "\"AwsServiceEvent\""),
				pattern.simple_expression("$.userIdentity.invokedBy", "NOT EXISTS", ""),
				pattern.simple_expression("$.userIdentity.type", "=", "\"Root\""),
			]),
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

# regal ignore:rule-length
test_fail if {
	# on the required filter alarm is not configured with sns topic
	# and on irrelevant filter an alaram is configured
	# so it should not pass
	eval_fail with input as rule_input([
		{
			"TrailInfo": {
				"Trail": {"IsMultiRegionTrail": true},
				"Status": {"IsLogging": true},
				"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
			},
			"MetricFilters": [{
				"FilterName": "filter_1",
				"FilterPattern": "{ $.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS && $.eventType != \"AwsServiceEvent\" }",
				"ParsedFilterPattern": pattern.complex_expression("&&", [
					pattern.simple_expression("$.userIdentity.type", "=", "\"Root\""),
					pattern.simple_expression("$.userIdentity.invokedBy", "NOT EXISTS", ""),
					pattern.simple_expression("$.eventType", "!=", "\"AwsServiceEvent\""),
				]),
			}],
			"MetricTopicBinding": {},
		},
		{
			"TrailInfo": {
				"Trail": {"IsMultiRegionTrail": true},
				"Status": {"IsLogging": true},
				"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
			},
			"MetricFilters": [{"FilterName": "filter_2", "FilterPattern": "{ pattern_not_matched: true }"}],
			"MetricTopicBinding": {"filter_2": ["arn:aws:...sns"]},
		},
	])

	# Different expression
	eval_fail with input as rule_input([
		{
			"TrailInfo": {
				"Trail": {"IsMultiRegionTrail": true},
				"Status": {"IsLogging": true},
				"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
			},
			"MetricFilters": [{
				"FilterName": "filter_1",
				"FilterPattern": "{ $.userIdentity.type = \"Non Root\" && $.userIdentity.invokedBy NOT EXISTS && $.eventType != \"AwsServiceEvent\" }",
				"ParsedFilterPattern": pattern.complex_expression("&&", [
					pattern.simple_expression("$.userIdentity.type", "=", "\"Non Root\""),
					pattern.simple_expression("$.userIdentity.invokedBy", "NOT EXISTS", ""),
					pattern.simple_expression("$.eventType", "!=", "\"AwsServiceEvent\""),
				]),
			}],
			"MetricTopicBinding": {},
		},
		{
			"TrailInfo": {
				"Trail": {"IsMultiRegionTrail": true},
				"Status": {"IsLogging": true},
				"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
			},
			"MetricFilters": [{"FilterName": "filter_2", "FilterPattern": "{ pattern_not_matched: true }"}],
			"MetricTopicBinding": {"filter_2": ["arn:aws:...sns"]},
		},
	])
}

rule_input(entry) = test_data.generate_monitoring_resources(entry)

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}
