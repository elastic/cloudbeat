package compliance.cis_aws.rules.cis_2_1_3

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input(false, false)
	eval_fail with input as rule_input(false, true)
	eval_fail with input as rule_input(true, false)
}

test_pass if {
	eval_pass with input as rule_input(true, true)
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_s3_bucket
	not_eval with input as test_data.generate_s3_bucket("Bucket", "", null, null, null, null)
}

rule_input(enabled, mfa_delete) := test_data.generate_s3_bucket("Bucket", "", null, test_data.generate_s3_bucket_versioning(enabled, mfa_delete), null, null)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
