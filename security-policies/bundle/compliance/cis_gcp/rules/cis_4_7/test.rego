package compliance.cis_gcp.rules.cis_4_7

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "cloud-compute"

subtype := "gcp-compute-disk"

test_violation {
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {}}, {})
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"diskEncryptionKey": {}}}, {})
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"diskEncryptionKey": {"sha256": ""}}}, {})
}

test_pass {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"diskEncryptionKey": {"sha256": "some_val"}}}, {})
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
