package compliance.cis_aws.rules.cis_3_5

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.compliance.policy.aws_config.ensure_config_enabled as audit
import data.lib.test
import future.keywords.if

finding := audit.finding

test_violation if {
	# single region, single recorder config
	eval_fail with input as rule_input(false, false)
	eval_fail with input as rule_input(true, false)
	eval_fail with input as rule_input(false, true)

	# multiple regions, multiple recorder config in each region
	eval_fail with input as test_data.aws_configservice_disabled_region_recorder

	# single region, no recorder config
	eval_fail with input as test_data.aws_configservice_empty_recorders
}

test_pass if {
	eval_pass with input as rule_input(true, true)
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_trail
}

rule_input(all_supported_enabled, include_global_resource_types_enabled) := test_data.generate_aws_configservice_with_resource([{"recorders": [test_data.generate_aws_configservice_recorder(all_supported_enabled, include_global_resource_types_enabled)]}])

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
