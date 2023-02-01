package compliance.cis_aws.rules.cis_4_4

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
		"MetricFilters": [{"FilterPattern": "{($.eventName=DeleteGroupPolicy)||($.eventName=DeleteRolePolicy)||($.eventNa me=DeleteUserPolicy)||($.eventName=PutGroupPolicy)||($.eventName=PutRolePolic y)||($.eventName=PutUserPolicy)||($.eventName=CreatePolicy)||($.eventName=Del etePolicy)||($.eventName=CreatePolicyVersion)||($.eventName=DeletePolicyVersi on)||($.eventName=AttachRolePolicy)||($.eventName=DetachRolePolicy)||($.event Name=AttachUserPolicy)||($.eventName=DetachUserPolicy)||($.eventName=AttachGr oupPolicy)||($.eventName=DetachGroupPolicy)}"}],
		"Topics": ["arn:aws:...sns"],
	}])
}

rule_input(entry) = test_data.generate_monitoring_resources(entry)

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
