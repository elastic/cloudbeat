package compliance.cis_aws.rules.cis_4_1

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

# regal ignore:rule-length
test_violation {
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
		"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"}],
	}])

	# No topics
	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true}],
		},
		"MetricTopicBinding": {},
		"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"}],
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
		"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])

	# The event selector does include management events
	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": false, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])

	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "WriteOnly"}],
		},
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
		"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"}],
	}])

	# Missing * wildcards
	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{"FilterPattern": "{ ($.errorCode = \"UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"}],
		"Topics": ["arn:aws:...sns"],
	}])

	eval_fail with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
		"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"}],
	}])
}

test_pass {
	eval_pass with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"}],
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
			"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") || ($.sourceIPAddress!=\"delivery.logs.amazonaws.com\") || ($.eventName!=\"HeadBucket\") }"}],
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

rule_input(entry) = test_data.generate_monitoring_resources(entry)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
