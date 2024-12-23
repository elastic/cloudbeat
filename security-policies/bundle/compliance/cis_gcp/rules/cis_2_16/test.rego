package compliance.cis_gcp.rules.cis_2_16

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "cloud-compute"

subtype := "gcp-compute-region-backend-service"

test_violation if {
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"protocol": "HTTPS"}}, null)
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"protocol": "HTTPS", "logConfig": {"enable": false}}}, null)
}

test_pass if {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"protocol": "HTTPS", "logConfig": {"enable": true}}}, null)
}

test_not_evaluated if {
	not_eval with input as test_data.generate_gcp_asset(type, subtype, {"data": {"protocol": "TCP", "logConfig": {"enable": true}}}, null)
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
