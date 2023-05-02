package compliance.cis_aws.rules.cis_1_16

import data.compliance.cis_aws.data_adapter
import data.lib.test

generate_input(statements) = {
	"subType": "aws-policy",
	"resource": {"document": {"Statement": statements}},
}

test_violation {
	# "Action" and "Resource" can be both lists and single strings
	eval_fail with input as generate_input([{
		"Action": ["*"],
		"Effect": "Allow",
		"Resource": ["*"],
	}])

	eval_fail with input as generate_input([{
		"Action": ["*"],
		"Effect": "Allow",
		"Resource": "*",
	}])

	eval_fail with input as generate_input([{
		"Action": "*",
		"Effect": "Allow",
		"Resource": ["*"],
	}])

	eval_fail with input as generate_input([{
		"Action": "*",
		"Effect": "Allow",
		"Resource": "*",
	}])

	# Multiple statements, only one is problematic
	eval_fail with input as generate_input([
		{
			"Action": "*",
			"Effect": "Allow",
			"Resource": ["my-resource"],
		},
		{
			"Action": "*",
			"Effect": "Allow",
			"Resource": "*",
		},
	])
}

test_pass {
	# No statements, no problems
	eval_pass with input as generate_input([])

	# Effect is not "Allow"
	eval_pass with input as generate_input([{
		"Action": ["*"],
		"Effect": "Some Other Effect",
		"Resource": ["*"],
	}])

	# Action is not *
	eval_pass with input as generate_input([{
		"Action": ["some-action"],
		"Effect": "Allow",
		"Resource": ["*"],
	}])

	# Resource is not *
	eval_pass with input as generate_input([{
		"Action": ["*"],
		"Effect": "Allow",
		"Resource": ["some-resource"],
	}])

	# Multiple statements but none is problematic by itself
	eval_pass with input as generate_input([
		{
			"Action": ["*"],
			"Effect": "Deny",
			"Resource": ["*"],
		},
		{
			"Action": ["*"],
			"Effect": "Allow",
			"Resource": ["some-resource"],
		},
		{
			"Action": ["some-action"],
			"Effect": "Allow",
			"Resource": "*",
		},
		{
			"Action": "*",
			"Effect": "Allow",
			"Resource": "some-resource",
		},
	])

	# Action contains wildchar but it doesn't match all resources
	eval_pass with input as generate_input([
		{
			"Action": "ec2:*",
			"Effect": "Allow",
			"Resource": "*",
		},
		{
			"Action": "*",
			"Effect": "Allow",
			"Resource": ["resource*", "*resource", "*resource*"],
		},
	])
}

test_not_evaluated {
	not_eval with input as {}
	not_eval with input as {"Statement": []} # No subType
}

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
