package compliance.cis_aws.rules.cis_5_4

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	# default security group with one inbound rule
	eval_fail with input as rule_input({"GroupName": "default", "IpPermissions": [{}]})

	# default security group with one outbound rule
	eval_fail with input as rule_input({"GroupName": "default", "IpPermissionsEgress": [{}]})
}

test_pass {
	# default security group with restricted inbound/outbound rules
	eval_pass with input as rule_input({"GroupName": "default", "IpPermissions": [], "IpPermissionsEgress": []})

	# non default security groups can have any rule
	eval_pass with input as rule_input({"GroupName": "custom", "IpPermissions": [{}], "IpPermissionsEgress": [{}]})
}

rule_input(entry) = test_data.generate_security_group(entry)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
