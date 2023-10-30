package compliance.cis_gcp.rules.cis_2_3

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "cloud-storage"

subtype := "gcp-logging-log-bucket"

test_violation {
	# no retention policy
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {}}, {})

	# retention policy is not locked
	eval_fail with input as test_data.generate_gcp_asset(type, subtype, {"data": {"retentionDays": 400}}, {})
}

test_pass {
	eval_pass with input as test_data.generate_gcp_asset(type, subtype, {"data": {"retentionDays": 400, "locked": true}}, {})
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
