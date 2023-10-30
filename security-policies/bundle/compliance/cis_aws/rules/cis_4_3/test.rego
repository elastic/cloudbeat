package compliance.cis_aws.rules.cis_4_3

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_pass {
	eval_pass with input as rule_input([{
		"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		},
		"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ $.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS && $.eventType != \"AwsServiceEvent\" }"}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

test_fail {
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
			"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ $.userIdentity.type = \"Root\" && $.userIdentity.invokedBy NOT EXISTS && $.eventType != \"AwsServiceEvent\" }"}],
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

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}
