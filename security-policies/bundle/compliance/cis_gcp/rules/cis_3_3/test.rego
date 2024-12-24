package compliance.cis_gcp.rules.cis_3_3

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "cloud-dns"

subtype := "gcp-dns-managed-zone"

test_violation if {
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"visibility": "PUBLIC"}}, null)
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"dnssecConfig": {"state": "OFF"}, "visibility": "PUBLIC"}}, null)
}

test_pass if {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"dnssecConfig": {"state": "ON"}, "visibility": "PUBLIC"}}, null)
}

test_not_evaluated if {
	not_eval with input as {}
	not_eval with input as test_data.generate_gcp_asset(type, subtype, {"data": {"dnssecConfig": {"state": "ON"}, "visibility": "PRIVATE"}}, null)
}

eval_fail if {
	test.assert_fail(finding) with data.benchmark_data_adapter as data_adapter
}

eval_pass if {
	test.assert_pass(finding) with data.benchmark_data_adapter as data_adapter
}

not_eval if {
	not finding with data.benchmark_data_adapter as data_adapter
}
