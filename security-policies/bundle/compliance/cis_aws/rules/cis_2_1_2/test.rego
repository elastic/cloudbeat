package compliance.cis_aws.rules.cis_2_1_2

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as test_data.generate_s3_bucket("Bucket", "", null, null, null, null)
	eval_fail with input as rule_input("Deny", "*", "s3:*", "true")
	eval_fail with input as rule_input("Deny", "*", "wrong", "false")
	eval_fail with input as rule_input("Deny", "wrong", "s3:*", "false")
	eval_fail with input as rule_input("Allow", "*", "s3:*", "false")
	eval_fail with input as test_data.generate_s3_bucket(
		"Bucket", "", [
			test_data.generate_s3_bucket_policy_statement("Allow", "*", "s3:*", "false"),
			test_data.generate_s3_bucket_policy_statement("Allow", "*", "s3:*", "false"),
		],
		null, null, null,
	)
}

test_pass if {
	eval_pass with input as rule_input("Deny", "*", "s3:*", "false")
	eval_pass with input as test_data.generate_s3_bucket(
		"Bucket", "", [
			test_data.generate_s3_bucket_policy_statement("Deny", "*", "s3:*", "false"),
			test_data.generate_s3_bucket_policy_statement("Allow", "*", "s3:*", "false"),
		],
		null, null, null,
	)
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_s3_bucket
	not_eval with input as test_data.s3_bucket_without_policy
}

rule_input(effect, principal, action, is_secure_transport) := test_data.generate_s3_bucket("Bucket", "", [test_data.generate_s3_bucket_policy_statement(effect, principal, action, is_secure_transport)], null, null, null)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
