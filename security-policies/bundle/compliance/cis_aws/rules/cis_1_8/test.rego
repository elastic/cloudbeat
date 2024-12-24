package compliance.cis_aws.rules.cis_1_8

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input(0, null)
	eval_fail with input as rule_input(13, null)
}

test_pass if {
	eval_pass with input as rule_input(14, null)
	eval_pass with input as rule_input(15, null)
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_pwd_policy
}

rule_input(pwd_len, reuse_count) := test_data.generate_password_policy(pwd_len, reuse_count)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
