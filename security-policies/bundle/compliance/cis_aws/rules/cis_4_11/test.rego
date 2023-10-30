package compliance.cis_aws.rules.cis_4_11

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
		"MetricFilters": [{"FilterName": "filter_1", "FilterPattern": "{ ($.eventName = CreateNetworkAcl) || ($.eventName = CreateNetworkAclEntry) || ($.eventName = DeleteNetworkAcl) || ($.eventName = DeleteNetworkAclEntry) || ($.eventName = ReplaceNetworkAclEntry) || ($.eventName = ReplaceNetworkAclAssociation) }"}],
		"MetricTopicBinding": {"filter_1": ["arn:aws:...sns"]},
	}])
}

rule_input(entry) = test_data.generate_monitoring_resources(entry)

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
