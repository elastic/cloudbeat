package compliance.cis_aws.rules.cis_1_20

import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

generate_input(analyzers, regions) := {
	"type": "identity-management",
	"subType": "aws-access-analyzers",
	"resource": {
		"Analyzers": analyzers,
		"Regions": regions,
	},
}

analyzer(arn, status, region) := {
	"Arn": arn,
	"CreatedAt": "2023-01-09T15:06:39Z",
	"Name": "Analyzer",
	"Status": status,
	"Type": "ACCOUNT",
	"Tags": {},
	"Region": region,
}

test_violation if {
	eval_fail with input as generate_input([], ["region-1"])
	eval_fail with input as generate_input([analyzer("some-arn", null, "region-1")], ["region-1"])
	eval_fail with input as generate_input([analyzer("some-arn", "FOO", "region-1")], ["region-1"])
	eval_fail with input as generate_input(
		[analyzer("some-arn", "ACTIVE", "region-1")],
		["region-1", "region-2"], # no analyzer in region-2
	)
	eval_fail with input as generate_input(
		[analyzer("some-arn", "ACTIVE", "region-1"), analyzer("invalid-status", "FOO", "region-2")],
		["region-1", "region-2"],
	)
}

test_pass if {
	# no regions, no problems
	eval_pass with input as generate_input(null, [])
	eval_pass with input as generate_input([], [])
	eval_pass with input as generate_input([analyzer("some-arn", "ACTIVE", "region-1")], ["region-1"])
	eval_pass with input as generate_input(
		[analyzer("some-arn", "ACTIVE", "region-1"), analyzer("invalid-status", "FOO", "region-1"), analyzer("some-other-arn", "ACTIVE", "region-2")],
		["region-1", "region-2"],
	)
}

test_not_evaluated if {
	not_eval with input as {}
	not_eval with input as {"resource": {"Analyzers": []}} # No subType
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
