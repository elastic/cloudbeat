package compliance.cis_gcp.rules.cis_3_5

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type = "cloud-dns"

subtype = "gcp-dns-managed-zone"

test_violation {
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"dnssecConfig": {"defaultKeySpecs": [{"algorithm": "RSASHA1", "keyType": "ZONE_SIGNING"}, {"algorithm": "RSASHA256", "keyType": "KEY_SIGNING"}]}, "visibility": "PUBLIC"}}, null)
}

test_pass {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"visibility": "PUBLIC"}}, null)
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"dnssecConfig": {"defaultKeySpecs": [{"algorithm": "RSASHA256", "keyType": "ZONE_SIGNING"}]}, "visibility": "PUBLIC"}}, null)
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"dnssecConfig": {"defaultKeySpecs": [{"algorithm": "RSASHA256", "keyType": "ZONE_SIGNING"}, {"algorithm": "RSASHA1", "keyType": "KEY_SIGNING"}]}, "visibility": "PUBLIC"}}, null)
}

test_not_evaluated {
	not_eval with input as {}
	not_eval with input as test_data.generate_gcp_asset(type, subtype, {"data": {"visibility": "FORWARDING"}}, null)
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
