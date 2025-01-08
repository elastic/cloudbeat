package compliance.cis_aws.rules.cis_1_14

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.compliance.lib.common
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input([{"active": true, "has_used": false, "rotation_date": common.past_date}], false, true, common.current_date, common.current_date)
	eval_fail with input as rule_input([{"active": true, "has_used": true, "rotation_date": "N/A"}], false, true, common.current_date, common.current_date)
}

test_pass if {
	eval_pass with input as rule_input([], false, false, "", "")
	eval_pass with input as rule_input([{"active": false, "has_used": false, "rotation_date": common.past_date}], false, true, common.current_date, common.current_date)
	eval_pass with input as rule_input([{"active": true, "has_used": false, "rotation_date": common.current_date, "last_access": "N/A"}], false, false, "", "")
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
