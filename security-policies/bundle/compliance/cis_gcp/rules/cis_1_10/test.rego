package compliance.cis_gcp.rules.cis_1_10

import data.cis_gcp.test_data
import data.compliance.lib.common
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input(null, "91d", common.current_date, {"state": "ENABLED"})
	eval_fail with input as rule_input(null, "89d", common.past_date, {"state": "ENABLED"})
	eval_fail with input as rule_input(null, "7776001s", common.current_date, {"state": "ENABLED"})
	eval_fail with input as rule_input(null, "7776000s", common.past_date, {"state": "ENABLED"})
	eval_fail with input as rule_input(null, "2160h", common.past_date, {"state": "ENABLED"})
	eval_fail with input as rule_input(null, "2161h", common.current_date, {"state": "ENABLED"})
}

test_pass if {
	eval_pass with input as rule_input(null, "90d", common.current_date, {"state": "ENABLED"})
	eval_pass with input as rule_input(null, "7776000s", common.current_date, {"state": "ENABLED"})
	eval_pass with input as rule_input(null, "2160h", common.current_date, {"state": "ENABLED"})
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
	not_eval with input as rule_input(["test.user@google.com"], "", "", {})
	not_eval with input as rule_input(["test.user@google.com"], "", "", {"state": "DISABLED"})
}

rule_input(members, rotationPeriod, nextRotationTime, primary) := test_data.generate_kms_resource(members, rotationPeriod, nextRotationTime, primary)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
