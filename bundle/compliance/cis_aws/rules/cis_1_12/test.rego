package compliance.cis_aws.rules.cis_1_12

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input([{"active": true}], false, true, "", "")
	eval_fail with input as rule_input([{"active": true}], false, true, test_data.past_date, "")
	eval_fail with input as rule_input([{"active": true}], false, true, "No_Information", test_data.past_date)
	eval_fail with input as rule_input([{"active": true, "last_access": test_data.future_date}], false, true, "", "")
	eval_fail with input as rule_input([{"active": true, "last_access": test_data.past_date}], false, true, test_data.past_date, "")
	eval_fail with input as rule_input([{"active": true, "has_used": false, "rotation_date": test_data.past_date}], false, true, test_data.future_date, test_data.future_date)
	eval_fail with input as rule_input([{"active": true, "last_access": test_data.past_date}, {"active": true, "last_access": test_data.future_date}], false, true, test_data.future_date, "")
}

test_pass {
	eval_pass with input as rule_input([{"active": false, "has_used": true, "last_access": test_data.past_date}], false, true, test_data.future_date, "")
	eval_pass with input as rule_input([{"active": true, "last_access": test_data.future_date, "has_used": true}], false, true, test_data.future_date, "")
	eval_pass with input as rule_input([{"active": true, "has_used": false, "rotation_date": test_data.future_date, "last_access": "N/A"}], false, false, "", "")
	eval_pass with input as rule_input([{"active": true, "has_used": true, "last_access": test_data.future_date}], false, true, "No_Information", test_data.future_date)
	eval_pass with input as rule_input([{"active": true, "last_access": test_data.future_date, "has_used": true}, {"active": true, "last_access": test_data.future_date, "has_used": true}], false, true, test_data.future_date, "")
	eval_pass with input as rule_input([{"active": true, "last_access": test_data.future_date, "has_used": true}, {"active": false, "last_access": test_data.past_date, "has_used": true}], false, true, test_data.future_date, "")
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
