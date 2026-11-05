package compliance.cis_aws.rules.cis_1_17

import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

generate_input(roles) := generate_input_with_arn("arn:aws:iam::aws:policy/AWSSupportAccess", roles)

govcloud_arn := "arn:aws-us-gov:iam::aws:policy/AWSSupportAccess"

generate_input_with_arn(arn, roles) := {
	"subType": "aws-policy",
	"resource": {
		"Arn": arn,
		"roles": roles,
	},
}

test_violation if {
	# "roles" field missing entirely
	eval_fail with input as {
		"subType": "aws-policy",
		"resource": {"Arn": "arn:aws:iam::aws:policy/AWSSupportAccess"},
	}

	# empty and null
	eval_fail with input as generate_input(null)
	eval_fail with input as generate_input([])
	eval_fail with input as generate_input_with_arn(govcloud_arn, [])

	# bad data
	eval_fail with input as generate_input([
		{"unexpected": "JSON"},
		{"some other": "value"},
	])
}

test_pass if {
	eval_pass with input as generate_input([{"RoleId": "some-id"}])
	eval_pass with input as generate_input([{"RoleId": "some-id"}, {"some other": "value"}])
	eval_pass with input as generate_input_with_arn(govcloud_arn, [{"RoleId": "some-id"}])
}

test_not_evaluated if {
	not_eval with input as {"resource": {}}
	not_eval with input as {"resource": {"roles": []}} # No subType
	not_eval with input as generate_input_with_arn("arn:aws-us-gov:iam::aws:policy/ReadOnlyAccess", [])
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
