package compliance.cis_aws.rules.cis_2_1_5

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test
import future.keywords.if

test_violation if {
	eval_fail with input as rule_input(false, false, false, false, false, false, false, false)

	# A single property is `false` in both account-level and bucket-specific public access block config
	eval_fail with input as rule_input(true, true, true, false, true, true, true, false)
	eval_fail with input as rule_input(true, true, false, true, true, true, false, true)
	eval_fail with input as rule_input(true, false, true, true, true, false, true, true)
	eval_fail with input as rule_input(false, true, true, true, false, true, true, true)

	# No public access block config
	eval_fail with input as test_data.generate_s3_bucket("Bucket", "", null, null, null, null)

	# Only bucket-level public access block config
	eval_fail with input as test_data.generate_s3_bucket("Bucket", "", null, null, test_data.generate_s3_public_access_block_configuration(true, false, true, true), null)

	# Only account-level public access block config
	eval_fail with input as test_data.generate_s3_bucket("Bucket", "", null, null, null, test_data.generate_s3_public_access_block_configuration(true, true, false, true))
}

test_pass if {
	eval_pass with input as rule_input(true, true, true, true, false, false, false, false)
	eval_pass with input as rule_input(false, false, false, false, true, true, true, true)
	eval_pass with input as rule_input(true, false, true, false, false, true, false, true)
	eval_pass with input as rule_input(false, true, false, true, true, false, true, false)

	# Only bucket-level public access block config
	eval_pass with input as test_data.generate_s3_bucket("Bucket", "", null, null, test_data.generate_s3_public_access_block_configuration(true, true, true, true), null)

	# Only account-level public access block config
	eval_pass with input as test_data.generate_s3_bucket("Bucket", "", null, null, null, test_data.generate_s3_public_access_block_configuration(true, true, true, true))
}

test_not_evaluated if {
	not_eval with input as test_data.not_evaluated_s3_bucket
}

rule_input(block_public_acls, block_public_policy, ignore_public_acls, restrict_public_buckets, account_block_public_acls, account_block_public_policy, account_ignore_public_acls, account_restrict_public_buckets) := test_data.generate_s3_bucket("Bucket", "", null, null, test_data.generate_s3_public_access_block_configuration(block_public_acls, block_public_policy, ignore_public_acls, restrict_public_buckets), test_data.generate_s3_public_access_block_configuration(account_block_public_acls, account_block_public_policy, account_ignore_public_acls, account_restrict_public_buckets))

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
