package compliance.cis_aws.rules.cis_5_1

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input({
		"CidrBlock": "0.0.0.0/0",
		"Egress": false,
		"PortRange": {
			"From": 0,
			"To": 1024,
		},
		"Protocol": "6",
		"RuleAction": "allow",
		"RuleNumber": 100,
	})

	eval_fail with input as rule_input({
		"CidrBlock": "0.0.0.0/0",
		"Egress": false,
		"PortRange": {
			"From": 40,
			"To": 80,
		},
		"Protocol": "6",
		"RuleAction": "allow",
		"RuleNumber": 100,
	})

	eval_fail with input as rule_input({
		"CidrBlock": "0.0.0.0/0",
		"Egress": false,
		"Protocol": "-1",
		"RuleAction": "allow",
		"RuleNumber": 32767,
	})

	eval_fail with input as rule_input({
		"CidrBlock": "0.0.0.0/0",
		"Egress": false,
		"PortRange": {
			"From": 3389,
			"To": 3390,
		},
		"Protocol": "6",
		"RuleAction": "allow",
		"RuleNumber": 100,
	})
}

test_pass if {
	eval_pass with input as rule_input({
		"CidrBlock": "0.0.0.0/0",
		"Egress": true,
		"Protocol": "-1",
		"RuleAction": "deny",
		"RuleNumber": 32767,
	})

	eval_pass with input as rule_input({
		"CidrBlock": "0.0.0.0/0",
		"Egress": false,
		"PortRange": {
			"From": 8080,
			"To": 8181,
		},
		"Protocol": "6",
		"RuleAction": "allow",
		"RuleNumber": 100,
	})

	eval_pass with input as rule_input({})
	eval_pass with input as rule_input({
		"CidrBlock": "0.0.0.0/0",
		"Egress": false,
		"PortRange": {
			"From": 40,
			"To": 41,
		},
		"Protocol": "6",
		"RuleAction": "allow",
		"RuleNumber": 100,
	})
}

rule_input(entry) := test_data.generate_nacl(entry)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}
