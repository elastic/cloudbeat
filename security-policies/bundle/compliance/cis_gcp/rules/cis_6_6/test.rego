package compliance.cis_gcp.rules.cis_6_6

import data.cis_gcp.test_data
import data.compliance.policy.gcp.data_adapter
import data.lib.test
import future.keywords.if

type := "cloud-database"

subtype := "gcp-sqladmin-instance"

test_violation if {
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

test_pass if {
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

test_not_evaluated if {
	not_eval with input as test_data.not_eval_resource
	not_eval with input as test_data.generate_gcp_asset(
		type, subtype,
		{"data": {"instanceType": "READ_REPLICA_INSTANCE"}},
		{},
	)
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
