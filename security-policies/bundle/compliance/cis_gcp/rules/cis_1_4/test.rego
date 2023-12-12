package compliance.cis_gcp.rules.cis_1_4

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	# fail if managed by user
	eval_fail with input as test_data.generate_iam_policy(["user:foo"], "role")
}

test_pass if {
	# pass if not managed by user
	eval_pass with input as test_data.generate_iam_policy(["foo"], "role")
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
	not_eval with input as test_data.no_policy_resource
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
