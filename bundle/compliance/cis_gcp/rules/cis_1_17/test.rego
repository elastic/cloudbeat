package compliance.cis_gcp.rules.cis_1_17

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "data-processing"

subtype := "gcp-dataproc-cluster"

test_violation {
	# doesn't have customer encryption key
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"config": {"encryptionConfig": {"gcePdKmsKeyName": null}}}},
		{},
	)
}

test_pass {
	# has a customer encryption key
	eval_pass with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"config": {"encryptionConfig": {"gcePdKmsKeyName": "projects/some_prj/locations/global/keyRings/some_keyring/cryptoKeys/some_key"}}}},
		{},
	)
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
}

eval_fail {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval {
	not finding with data.benchmark_data_adapter as data_adapter
}
