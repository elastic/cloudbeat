package compliance.cis_aws.rules.cis_5_2

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input({"IpPermissions": [{
		"FromPort": 443,
		"IpProtocol": "tcp",
		"IpRanges": [{"CidrIp": "0.0.0.0/0"}],
		"Ipv6Ranges": [],
		"PrefixListIds": [],
		"ToPort": 443,
	}]})

	# "FromPort" and "ToPort" fields are not set in a security group rule, it means that the rule applies to all ports.
	eval_fail with input as rule_input({"IpPermissions": [{
		"IpProtocol": "tcp",
		"IpRanges": [{"CidrIp": "0.0.0.0/0"}],
		"Ipv6Ranges": [],
		"PrefixListIds": [],
		"UserIdGroupPairs": [],
	}]})
}

test_pass if {
	# IpRanges empty array
	# no inbound traffic is allowed to reach the resources associated with that security group
	eval_pass with input as rule_input({"IpPermissions": [{
		"FromPort": 443,
		"IpProtocol": "tcp",
		"IpRanges": [],
		"Ipv6Ranges": [],
		"PrefixListIds": [],
		"ToPort": 443,
		"UserIdGroupPairs": [],
	}]})

	# IpRanges with CiderIP different from 0.0.0.0/0 is OK
	eval_pass with input as rule_input({"IpPermissions": [{
		"FromPort": 22,
		"IpProtocol": "tcp",
		"IpRanges": [{"CidrIp": "31.154.188.106/32"}],
		"Ipv6Ranges": [],
		"PrefixListIds": [],
		"ToPort": 22,
		"UserIdGroupPairs": [],
	}]})
}

rule_input(entry) := test_data.generate_security_group(entry)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
