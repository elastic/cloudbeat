package compliance.cis_gcp.rules.cis_2_3

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "cloud-storage"

subtype := "gcp-logging-log-bucket"

test_violation if {
	# no retention policy
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {}}, {})

	# retention policy is not locked
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"retentionDays": 400}}, {})
}

test_pass if {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"retentionDays": 400, "locked": true}}, {})
}

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
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
