package compliance.cis_aws.rules.cis_3_6

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input(false, null, null, false, null)
	eval_fail with input as rule_input(false, null, null, null, null)
	eval_fail with input as rule_input(false, null, null, "", null)
}

test_pass {
	eval_pass with input as rule_input(false, null, null, true, null)
}

test_not_evaluated {
	not_eval with input as test_data.not_evaluated_trail
}

rule_input(is_log_validation_enabled, cloudwatch_log_group_arn, log_delivery_time, is_bucket_logging_enabled, kms_key_id) = test_data.generate_enriched_trail(is_log_validation_enabled, cloudwatch_log_group_arn, log_delivery_time, is_bucket_logging_enabled, kms_key_id)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
