package compliance.cis_aws.rules.cis_1_15

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input([{"inline_policy"}], [])
	eval_fail with input as rule_input([], [{"attached_policy"}])
	eval_fail with input as rule_input([{"inline_policy"}], [{"attached_policy"}])
}

test_pass if {
	eval_pass with input as rule_input([], [])
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_iam_user
}

rule_input(inline_policies, attached_policies) := test_data.generate_iam_user_with_policies(inline_policies, attached_policies)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
