package compliance.cis_gcp.rules.cis_5_2

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input(false)
	eval_fail with input as rule_input(null)
}

test_pass if {
	eval_pass with input as rule_input(true)
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
}

rule_input(isBucketLevelAccessEnabled) := test_data.generate_gcs_resource([], isBucketLevelAccessEnabled)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
