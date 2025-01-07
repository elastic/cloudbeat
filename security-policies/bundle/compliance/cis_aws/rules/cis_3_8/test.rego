package compliance.cis_aws.rules.cis_3_8

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.compliance.policy.aws_kms.ensure_symmetric_key_rotation_enabled as audit
import data.lib.test
import future.keywords.if

finding := audit.finding

test_violation if {
	eval_fail with input as rule_input(false)
}

test_pass if {
	eval_pass with input as rule_input(true)
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_trail
}

rule_input(symmetric_default_enabled) := test_data.generate_kms_resource(symmetric_default_enabled)

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
