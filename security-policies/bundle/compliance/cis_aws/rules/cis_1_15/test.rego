package compliance.cis_aws.rules.cis_1_15

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input([{"inline_policy"}], [])
	eval_fail with input as rule_input([], [{"attached_policy"}])
	eval_fail with input as rule_input([{"inline_policy"}], [{"attached_policy"}])
}

test_pass {
	eval_pass with input as rule_input([], [])
}

test_not_evaluated {
	not_eval with input as test_data.not_evaluated_iam_user
}

rule_input(inline_policies, attached_policies) = test_data.generate_iam_user_with_policies(inline_policies, attached_policies)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
