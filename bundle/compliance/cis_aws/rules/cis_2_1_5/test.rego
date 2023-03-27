package compliance.cis_aws.rules.cis_2_1_5

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	# All properties false
	eval_fail with input as rule_input(false, false, false, false)

	# 3 properties false
	eval_fail with input as rule_input(true, false, false, false)
	eval_fail with input as rule_input(false, true, false, false)
	eval_fail with input as rule_input(false, false, true, false)
	eval_fail with input as rule_input(false, false, false, true)

	# 2 properties false
	eval_fail with input as rule_input(true, true, false, false)
	eval_fail with input as rule_input(true, false, true, false)
	eval_fail with input as rule_input(true, false, false, true)
	eval_fail with input as rule_input(false, true, true, false)
	eval_fail with input as rule_input(false, true, false, true)
	eval_fail with input as rule_input(false, false, true, true)

	# 1 property false
	eval_fail with input as rule_input(false, true, true, true)
	eval_fail with input as rule_input(true, false, true, true)
	eval_fail with input as rule_input(true, true, false, true)
	eval_fail with input as rule_input(true, true, true, false)
}

test_pass {
	eval_pass with input as rule_input(true, true, true, true)
}

test_not_evaluated {
	not_eval with input as test_data.not_evaluated_s3_bucket
	not_eval with input as test_data.generate_s3_bucket("Bucket", "", null, null, null)
}

rule_input(block_public_acls, block_public_policy, ignore_public_acls, restrict_public_buckets) = test_data.generate_s3_bucket("Bucket", "", null, null, test_data.generate_s3_public_access_block_configuration(block_public_acls, block_public_policy, ignore_public_acls, restrict_public_buckets))

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
