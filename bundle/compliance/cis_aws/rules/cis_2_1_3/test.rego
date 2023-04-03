package compliance.cis_aws.rules.cis_2_1_3

import data.cis_aws.test_data
import data.compliance.cis_aws.data_adapter
import data.lib.test

test_violation {
	eval_fail with input as rule_input(false, false)
	eval_fail with input as rule_input(false, true)
	eval_fail with input as rule_input(true, false)
}

test_pass {
	eval_pass with input as rule_input(true, true)
}

test_not_evaluated {
	not_eval with input as test_data.not_evaluated_s3_bucket
	not_eval with input as test_data.generate_s3_bucket("Bucket", "", null, null, null, null)
}

rule_input(enabled, mfa_delete) = test_data.generate_s3_bucket("Bucket", "", null, test_data.generate_s3_bucket_versioning(enabled, mfa_delete), null, null)

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
