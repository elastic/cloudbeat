package compliance.cis_aws.rules.cis_2_1_1

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input("my bucket", "")
	eval_fail with input as rule_input("my bucket", "FakeAlgorithm")
	eval_fail with input as rule_input("my bucket", "NoEncryption")
}

test_pass {
	eval_pass with input as rule_input("my bucket", "AES256")
	eval_pass with input as rule_input("my bucket", "aws:kms")
}

test_not_evaluated {
	not_eval with input as test_data.not_evaluated_s3_bucket
	not_eval with input as rule_input("my bucket", null)
}

rule_input(name, sse_algorithm) = test_data.generate_s3_bucket(name, sse_algorithm, null, null, null, null)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
