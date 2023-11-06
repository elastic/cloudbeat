package compliance.cis_aws.rules.cis_1_13

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input([{"active": true, "has_used": true}, {"active": true, "has_used": false}], false, true, "", "")
}

test_pass {
	eval_pass with input as rule_input([], false, null, "", "")
	eval_pass with input as rule_input([{"active": true}], false, null, "", "")
	eval_pass with input as rule_input([{"active": false}, {"active": false}], false, null, "", "")
	eval_pass with input as rule_input([{"active": true}, {"active": false}], false, null, "", "")
}

test_not_evaluated {
	not_eval with input as test_data.not_evaluated_iam_user
}

rule_input(access_keys, mfa_active, password_enabled, last_access, password_last_changed) = test_data.generate_iam_user(access_keys, mfa_active, password_enabled, last_access, password_last_changed)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
