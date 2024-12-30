package compliance.cis_aws.rules.cis_1_7

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.compliance.lib.common
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input([{"active": true, "last_access": common.current_date}], true, common.past_date, [])
	eval_fail with input as rule_input([{"active": true, "last_access": common.past_date}], true, common.current_date, [])
	eval_fail with input as rule_input([{"active": true, "last_access": common.current_date}], true, common.current_date, [])
	eval_fail with input as rule_input([{"active": true, "last_access": common.current_date}, {"active": true, "last_access": common.past_date}], true, common.current_date, [])
}

test_pass if {
	eval_pass with input as rule_input([], true, common.past_date, [])
	eval_pass with input as rule_input([{"active": true, "last_access": common.past_date}], true, common.past_date, [])
	eval_pass with input as rule_input([{"active": false, "last_access": common.current_date}], true, common.past_date, [])
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_iam_user
}

rule_input(access_keys, mfa_active, last_access, mfa_devices) := test_data.generate_root_user(access_keys, mfa_active, last_access, mfa_devices)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
