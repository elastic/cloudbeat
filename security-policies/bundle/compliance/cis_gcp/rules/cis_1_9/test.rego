package compliance.cis_gcp.rules.cis_1_9

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input(["allUsers", "test.user@google.com"], "", "", {"state": "ENABLED"})
	eval_fail with input as rule_input(["allAuthenticatedUsers"], "", "", {"state": "ENABLED"})
	eval_fail with input as rule_input(["allUsers", "allAuthenticatedUsers"], "", "", {"state": "ENABLED"})
}

test_pass if {
	eval_pass with input as {"subType": "gcp-cloudkms-crypto-key", "resource": {}} # a resource without an iam_policy
	eval_pass with input as rule_input([], "", "", {"state": "ENABLED"})
	eval_pass with input as rule_input(["test.user@google.com"], "", "", {"state": "ENABLED"})
	eval_pass with input as rule_input(["test.user@google.com", "test.user2@google.com"], "", "", {"state": "ENABLED"})
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
}

rule_input(members, nextRotationTime, rotationPeriod, primary) := test_data.generate_kms_resource(members, nextRotationTime, rotationPeriod, primary)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
