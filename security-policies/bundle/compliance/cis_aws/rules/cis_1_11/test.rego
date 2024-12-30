package compliance.cis_aws.rules.cis_1_11

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input([{"active": true, "has_used": false}], null, true, "", "")
	eval_fail with input as rule_input([{"active": true, "has_used": true}, {"active": true, "has_used": false}], null, true, "", "")
}

test_pass if {
	eval_pass with input as rule_input(null, null, false, "", "")
	eval_pass with input as rule_input(null, null, true, "", "")
	eval_pass with input as rule_input([], null, true, "", "")
	eval_pass with input as rule_input([{"active": true, "has_used": true}], null, true, "", "")
	eval_pass with input as rule_input([{"active": false, "has_used": true}], null, true, "", "")
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_iam_user
}

rule_input(access_keys, mfa_active, password_enabled, last_access, password_last_changed) := test_data.generate_iam_user(access_keys, mfa_active, password_enabled, last_access, password_last_changed)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
