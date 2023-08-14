package compliance.cis_gcp.rules.cis_6_6

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test

type := "cloud-database"

subtype := "gcp-sqladmin-instance"

test_violation {
	# fail when some ipAddresses is not set to private
	eval_fail with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {
			"instanceType": "CLOUD_SQL_INSTANCE",
			"backendType": "SECOND_GEN",
			"ipAddresses": [{"type": "PRIMARY"}, {"type": "PRIVATE"}],
		}},
		{},
	)
}

test_pass {
	# pass when ipAddresses is set to private
	eval_pass with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {
			"instanceType": "CLOUD_SQL_INSTANCE",
			"backendType": "SECOND_GEN",
			"ipAddresses": [{"type": "PRIVATE"}],
		}},
		{},
	)
}

test_not_evaluated {
	not_eval with input as test_data.not_eval_resource
	not_eval with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"instanceType": "READ_REPLICA_INSTANCE"}},
		{},
	)
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
