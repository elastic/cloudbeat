package compliance.cis_aws.rules.cis_3_1

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# No items
	eval_fail with input as rule_input([])

	# No trails with "IsMultiRegionTrail=true"
	eval_fail with input as rule_input([{"TrailInfo": {
		"Trail": {"IsMultiRegionTrail": false},
		"Status": {"IsLogging": true},
		"EventSelectors": [{"IncludeManagementEvents": true}],
	}}])

	# No "IsLogging=true"
	eval_fail with input as rule_input([{"TrailInfo": {
		"Trail": {"IsMultiRegionTrail": true},
		"Status": {"IsLogging": false},
		"EventSelectors": [{"IncludeManagementEvents": true}],
	}}])

	# The event selector does include management events
	eval_fail with input as rule_input([{"TrailInfo": {
		"Trail": {"IsMultiRegionTrail": true},
		"Status": {"IsLogging": true},
		"EventSelectors": [{"IncludeManagementEvents": false, "ReadWriteType": "All"}],
	}}])

	eval_fail with input as rule_input([{"TrailInfo": {
		"Trail": {"IsMultiRegionTrail": true},
		"Status": {"IsLogging": true},
		"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "WriteOnly"}],
	}}])
}

test_pass if {
	eval_pass with input as rule_input([{"TrailInfo": {
		"Trail": {"IsMultiRegionTrail": true},
		"Status": {"IsLogging": true},
		"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
	}}])

	# at least one is matching
	eval_pass with input as rule_input([
		{"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": true},
			"Status": {"IsLogging": true},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "All"}],
		}},
		{"TrailInfo": {
			"Trail": {"IsMultiRegionTrail": false},
			"Status": {"IsLogging": false},
			"EventSelectors": [{"IncludeManagementEvents": true, "ReadWriteType": "WriteOnly"}],
		}},
	])
}

rule_input(entry) := test_data.generate_monitoring_resources(entry)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
