package compliance.cis_aws.rules.cis_2_1_1

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input("my bucket", "")
	eval_fail with input as rule_input("my bucket", "FakeAlgorithm")
	eval_fail with input as rule_input("my bucket", "NoEncryption")
}

test_pass if {
	eval_pass with input as rule_input("my bucket", "AES256")
	eval_pass with input as rule_input("my bucket", "aws:kms")
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_s3_bucket
	not_eval with input as rule_input("my bucket", null)
}

rule_input(name, sse_algorithm) := test_data.generate_s3_bucket(name, sse_algorithm, null, null, null, null)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
