package compliance.cis_aws.rules.cis_4_1

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.compliance.policy.aws_cloudtrail.pattern
import data.lib.test
import future.keywords.if

equal_pattern := pattern.complex_expression("||", [
	pattern.simple_expression("$.errorCode", "=", "\"*UnauthorizedOperation\""),
	pattern.simple_expression("$.errorCode", "=", "\"AccessDenied*\""),
	pattern.simple_expression("$.sourceIPAddress", "!=", "\"delivery.logs.amazonaws.com\""),
	pattern.simple_expression("$.eventName", "!=", "\"HeadBucket\""),
])

valid_different_order_pattern := pattern.complex_expression("||", [
	pattern.simple_expression("$.eventName", "!=", "\"HeadBucket\""),
	pattern.simple_expression("$.errorCode", "=", "\"AccessDenied*\""),
	pattern.simple_expression("\"delivery.logs.amazonaws.com\"", "!=", "$.sourceIPAddress"),
	pattern.simple_expression("$.errorCode", "=", "\"*UnauthorizedOperation\""),
])

missing_wildcard_pattern := pattern.complex_expression("||", [
	pattern.simple_expression("$.errorCode", "=", "\"UnauthorizedOperation\""),
	pattern.simple_expression("$.errorCode", "=", "\"AccessDenied*\""),
	pattern.simple_expression("$.sourceIPAddress", "!=", "\"delivery.logs.amazonaws.com\""),
	pattern.simple_expression("$.eventName", "!=", "\"HeadBucket\""),
])

# regal ignore:rule-length
test_violation if {
	# No items
	eval_fail with input as rule_input([])

	# No trails with "IsMultiRegionTrail=true"
	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": false},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true}],
		},
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }",
			"ParsedFilterPattern": equal_pattern,
		}],
	}])

	# No topics
	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true}],
		},
		"MetricTopicBinding": {},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }",
			"ParsedFilterPattern": equal_pattern,
		}],
	}])

	# No "AuthorizationFilterExists=true"
	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true}],
		},
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
		"MetricFilters": [],
	}])

	# No "IsLogging=true"
	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": false},
			"EventSelectors": [{"IncludeManagementEvents": true}],
		},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }",
			"ParsedFilterPattern": equal_pattern,
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])

	# The event selector does include management events
	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": false, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }",
			"ParsedFilterPattern": equal_pattern,
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])

	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "WriteOnly"}],
		},
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }",
			"ParsedFilterPattern": equal_pattern,
		}],
	}])

	# Missing * wildcards
	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.errorCode = \"UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }",
			"ParsedFilterPattern": missing_wildcard_pattern,
		}],
		"Topics": ["arn:aws:...sns"],
	}])
}

# regal ignore:rule-length
test_pass if {
	eval_pass with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }",
			"ParsedFilterPattern": equal_pattern,
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])

	eval_pass with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{
			"FilterName": "filter_1",
			"ParsedFilterPattern": valid_different_order_pattern,
		}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])

	# at least one is matching
	eval_pass with input as rule_input([
		{
			"TrailInfo": {
				"Trail": {"IsMultiRegionTrail": true},
				"Status": {"IsLogging": true},
				"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
			},
			"MetricFilters": [{
				"FilterName": "filter_1",
				"FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }",
				"ParsedFilterPattern": equal_pattern,
			}],
			"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
		},
		{
			"TrailInfo": {
				"Trail": {"IsMultiRegionTrail": false},
				"Status": {"IsLogging": false},
				"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "WriteOnly"}],
			},
			"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
			"MetricFilters": [],
		},
	])
}

rule_input(entry) := test_data.generate_monitoring_resources(entry)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
